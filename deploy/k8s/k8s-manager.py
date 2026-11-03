#!/usr/bin/env python3
"""
Kubernetes Cluster & Workload Orchestrator for Knowledge Graph Platform
Manages:
1. Stopping & removing Docker Compose containers and freeing image disk space.
2. Local KinD (Kubernetes-in-Docker) cluster provisioning with host port mappings.
3. Applying base Kustomization manifests (deploy/k8s/base).
4. Verifying workload rollouts, services, and pods.
"""

import argparse
import os
import shutil
import subprocess
import sys
import time

REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "../.."))
K8S_BASE_DIR = os.path.join(REPO_ROOT, "deploy/k8s/base")
KIND_CONFIG_PATH = os.path.join(REPO_ROOT, "deploy/k8s/kind-cluster-config.yaml")
COMPOSE_FILE = os.path.join(REPO_ROOT, "deploy/docker/compose/docker-compose.infra.yml")
DOCKERFILE_SERVICE = os.path.join(REPO_ROOT, "deploy/docker/service/Dockerfile")
DOCKERFILE_WORKER = os.path.join(REPO_ROOT, "deploy/docker/service/Dockerfile.python-ingestion")

DOCKER_BIN = "/snap/bin/docker" if os.path.exists("/snap/bin/docker") else "docker"
KIND_BIN = "/home/btpl-lap-22/.local/bin/kind" if os.path.exists("/home/btpl-lap-22/.local/bin/kind") else "kind"
KUBECTL_BIN = "/home/btpl-lap-22/google-cloud-sdk/bin/kubectl" if os.path.exists("/home/btpl-lap-22/google-cloud-sdk/bin/kubectl") else "kubectl"

CLUSTER_NAME = "kg-platform-cluster"
APP_IMAGE = "ghcr.io/knowledge-graph-platform/knowledge-graph-platform:1.0.0"
GRAFANA_IMAGE = "grafana/grafana:11.2.0"


def run_command(args, check=True, capture=False, env=None):
    merged_env = os.environ.copy()
    if env:
        merged_env.update(env)
    return subprocess.run(
        args,
        cwd=REPO_ROOT,
        check=check,
        text=True,
        stdout=subprocess.PIPE if capture else None,
        stderr=subprocess.PIPE if capture else None,
        env=merged_env,
    )


def clean_docker():
    print("[INFO] Stopping and deleting all Docker Compose containers...")
    run_command([DOCKER_BIN, "compose", "-f", COMPOSE_FILE, "down", "--remove-orphans"], check=False)

    print("[INFO] Removing untagged and stopped container artifacts...")
    run_command([DOCKER_BIN, "container", "prune", "-f"], check=False)

    print("[SUCCESS] Docker container environment cleaned.")


def create_kind_cluster():
    print(f"[INFO] Checking KinD cluster status for '{CLUSTER_NAME}'...")
    res = run_command([KIND_BIN, "get", "clusters"], capture=True)
    if CLUSTER_NAME not in res.stdout:
        print(f"[INFO] Creating KinD cluster '{CLUSTER_NAME}' with ingress port forwarding...")
        
        kind_config = """kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 80
    hostPort: 8000
    protocol: TCP
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
  - containerPort: 3000
    hostPort: 3000
    protocol: TCP
"""
        with open(KIND_CONFIG_PATH, "w") as f:
            f.write(kind_config)
            
        run_command([KIND_BIN, "create", "cluster", "--name", CLUSTER_NAME, "--config", KIND_CONFIG_PATH])
    else:
        print(f"[INFO] Cluster '{CLUSTER_NAME}' already active.")


def build_and_load_images():
    print(f"[INFO] Building Go platform image: {APP_IMAGE}...")
    run_command([DOCKER_BIN, "build", "-t", APP_IMAGE, "-f", DOCKERFILE_SERVICE, "."])

    tmp_dir = os.path.expanduser("~/docker-tmp")
    os.makedirs(tmp_dir, exist_ok=True)

    print(f"[INFO] Pulling and loading Grafana image: {GRAFANA_IMAGE}...")
    run_command([DOCKER_BIN, "pull", GRAFANA_IMAGE])
    run_command(
        [KIND_BIN, "load", "docker-image", GRAFANA_IMAGE, "--name", CLUSTER_NAME],
        env={"TMPDIR": tmp_dir},
    )

    print(f"[INFO] Loading image {APP_IMAGE} into KinD cluster '{CLUSTER_NAME}'...")
    run_command(
        [KIND_BIN, "load", "docker-image", APP_IMAGE, "--name", CLUSTER_NAME],
        env={"TMPDIR": tmp_dir},
    )
    print(f"[SUCCESS] Image {APP_IMAGE} loaded into KinD cluster.")


def deploy_k8s():
    print("[INFO] Applying Kubernetes manifests from deploy/k8s/base...")
    run_command([KUBECTL_BIN, "apply", "-k", K8S_BASE_DIR])

    print("[INFO] Waiting for resources to register in Kubernetes...")
    time.sleep(3)
    
    print("\n[DEPLOYED KUBERNETES RESOURCES]")
    run_command([KUBECTL_BIN, "get", "all", "-l", "app.kubernetes.io/part-of=knowledge-graph-system"])


def restart_workloads():
    print("[INFO] Restarting deployment/knowledge-graph-platform...")
    run_command([KUBECTL_BIN, "rollout", "restart", "deployment/knowledge-graph-platform"])
    print("[INFO] Waiting for rollout to finish...")
    run_command([KUBECTL_BIN, "rollout", "status", "deployment/knowledge-graph-platform", "--timeout=120s"])


def port_forward_grafana(port="13000"):
    print(f"[INFO] Forwarding Grafana to http://127.0.0.1:{port} (auto-reconnecting)...")
    while True:
        try:
            subprocess.run([KUBECTL_BIN, "port-forward", "svc/grafana", f"{port}:3000", "--address", "0.0.0.0"])
        except KeyboardInterrupt:
            print("\n[INFO] Stopped port-forward.")
            break
        except Exception as e:
            print(f"[WARN] Port-forward disconnected: {e}. Reconnecting in 2s...")
        time.sleep(2)


def verify_endpoints():
    print("\n=== VERIFYING ENDPOINTS VIA KUBECTL PROXY / PORT-FORWARD ===")
    port = "18000"
    pf_proc = subprocess.Popen(
        [KUBECTL_BIN, "port-forward", "svc/knowledge-graph-platform", f"{port}:8000"],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    time.sleep(2)
    try:
        import urllib.request
        for path in ["/healthz", "/ready"]:
            url = f"http://127.0.0.1:{port}{path}"
            req = urllib.request.urlopen(url, timeout=5)
            content = req.read().decode("utf-8")
            print(f"[CHECK] {url} -> HTTP {req.status}: {content.strip()}")
            assert req.status == 200, f"Endpoint {url} returned non-200"
        print("[SUCCESS] Platform endpoints healthy and verified.")
    finally:
        pf_proc.terminate()
        pf_proc.wait()

    # Verify Grafana endpoint via its own test forwarder
    g_port = "13000"
    g_pf = subprocess.Popen(
        [KUBECTL_BIN, "port-forward", "svc/grafana", f"{g_port}:3000"],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    time.sleep(2)
    try:
        import urllib.request
        url = f"http://127.0.0.1:{g_port}/api/health"
        req = urllib.request.urlopen(url, timeout=5)
        content = req.read().decode("utf-8")
        print(f"[CHECK] Grafana {url} -> HTTP {req.status}: {content.strip()}")
        print(f"[SUCCESS] Grafana verified successfully at http://127.0.0.1:{g_port}")
    except Exception as e:
        print(f"[ERROR] Grafana healthcheck failed: {e}")
    finally:
        g_pf.terminate()
        g_pf.wait()



def status_k8s():
    print("\n=== POD STATUS ===")
    run_command([KUBECTL_BIN, "get", "pods", "-o", "wide"])
    print("\n=== SERVICES ===")
    run_command([KUBECTL_BIN, "get", "svc"])
    print("\n=== AUTOSCALERS & PDB ===")
    run_command([KUBECTL_BIN, "get", "hpa,pdb"])


def inject_telemetry():
    print("\n=== INJECTING TELEMETRY (TRACES & LOGS) INTO KUBERNETES OBS STACK ===")
    import json
    import random
    import time
    import urllib.request

    trace_id = "".join([random.choice("0123456789abcdef") for _ in range(32)])
    span_id = "".join([random.choice("0123456789abcdef") for _ in range(16)])
    now_nano = str(time.time_ns())

    # 1. Forward Loki (3100) and Tempo (4318) for injection
    print("[INFO] Establishing temporary port-forwards for injection...")
    loki_pf = subprocess.Popen(
        [KUBECTL_BIN, "port-forward", "svc/loki", "13100:3100"],
    )
    tempo_pf = subprocess.Popen(
        [KUBECTL_BIN, "port-forward", "svc/tempo", "14318:4318"],
    )
    time.sleep(4)

    try:
        # Ingest Log into Loki
        log_payload = {
            "streams": [
                {
                    "stream": {
                        "job": "knowledge-graph-platform",
                        "service_name": "knowledge-graph-platform",
                        "environment": "production",
                        "level": "info",
                    },
                    "values": [
                        [
                            now_nano,
                            f'{{"timestamp":"{time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())}","trace_id":"{trace_id}","span_id":"{span_id}","message":"Executing ontology transformation job: ont-finance-v1","service":"kg-platform","status":"processed"}}',
                        ]
                    ],
                }
            ]
        }
        loki_req = urllib.request.Request(
            "http://127.0.0.1:13100/loki/api/v1/push",
            data=json.dumps(log_payload).encode("utf-8"),
            headers={"Content-Type": "application/json"},
        )
        with urllib.request.urlopen(loki_req, timeout=5) as resp:
            print(f"[SUCCESS] Ingested log to Loki (HTTP {resp.status}) with trace_id: {trace_id}")

        # Ingest Trace into Tempo via OTLP HTTP
        now_unix_nano = time.time_ns()
        start_nano = str(now_unix_nano - 50_000_000)
        end_nano = str(now_unix_nano)

        trace_payload = {
            "resourceSpans": [
                {
                    "resource": {
                        "attributes": [
                            {"key": "service.name", "value": {"stringValue": "knowledge-graph-platform"}},
                            {"key": "service.version", "value": {"stringValue": "1.0.0"}},
                            {"key": "deployment.environment", "value": {"stringValue": "production"}},
                        ]
                    },
                    "scopeSpans": [
                        {
                            "scope": {"name": "kg-query-engine"},
                            "spans": [
                                {
                                    "traceId": trace_id,
                                    "spanId": span_id,
                                    "name": "OntologyTransform::Execute",
                                    "kind": 1,
                                    "startTimeUnixNano": start_nano,
                                    "endTimeUnixNano": end_nano,
                                    "attributes": [
                                        {"key": "graph.schema", "value": {"stringValue": "RDF/OWL"}},
                                        {"key": "entities.transformed", "value": {"intValue": 420}},
                                        {"key": "http.status_code", "value": {"intValue": 200}},
                                    ],
                                    "status": {"code": 1},
                                }
                            ],
                        }
                    ],
                }
            ]
        }

        tempo_req = urllib.request.Request(
            "http://127.0.0.1:14318/v1/traces",
            data=json.dumps(trace_payload).encode("utf-8"),
            headers={"Content-Type": "application/json"},
        )
        with urllib.request.urlopen(tempo_req, timeout=5) as resp:
            print(f"[SUCCESS] Ingested trace to Tempo (HTTP {resp.status}) with Trace ID: {trace_id}")

        print(f"\n[EXPLORE LINKS]")
        print(f"  Trace in Tempo: http://127.0.0.1:13000/explore?schemaVersion=1&panes=%7B%22s0v%22:%7B%22datasource%22:%22P214B5B846CF3925F%22,%22queries%22:%5B%7B%22query%22:%22{trace_id}%22,%22queryType%22:%22traceId%22%7D%5D%7D%7D")
        print(f"  Logs in Loki:   http://127.0.0.1:13000/explore?schemaVersion=1&panes=%7B%22s0v%22:%7B%22datasource%22:%22P8E80F9AEF21F6940%22,%22queries%22:%5B%7B%22expr%22:%22%7Bjob%3D%5C%22knowledge-graph-platform%5C%22%7D%22%7D%5D%7D%7D")
    finally:
        loki_pf.terminate()
        tempo_pf.terminate()
        loki_pf.wait()
        tempo_pf.wait()


def main():
    parser = argparse.ArgumentParser(description="Knowledge Graph Platform Kubernetes Orchestrator")
    parser.add_argument(
        "action",
        choices=["clean-docker", "setup-cluster", "build-load", "deploy", "restart", "verify", "inject-telemetry", "port-forward-grafana", "status", "all"],
        help="Action to perform",
    )
    args = parser.parse_args()

    if args.action == "clean-docker":
        clean_docker()
    elif args.action == "setup-cluster":
        create_kind_cluster()
    elif args.action == "build-load":
        build_and_load_images()
    elif args.action == "deploy":
        deploy_k8s()
    elif args.action == "restart":
        restart_workloads()
    elif args.action == "verify":
        verify_endpoints()
    elif args.action == "inject-telemetry":
        inject_telemetry()
    elif args.action == "port-forward-grafana":
        port_forward_grafana()
    elif args.action == "status":
        status_k8s()
    elif args.action == "all":
        clean_docker()
        create_kind_cluster()
        build_and_load_images()
        deploy_k8s()
        restart_workloads()
        verify_endpoints()
        status_k8s()


if __name__ == "__main__":
    main()
