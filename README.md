# Hello Politecnico Scheduler

Out-of-tree Kubernetes custom scheduler built on the official
[Scheduling Framework](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/)
(`k8s.io/kubernetes/cmd/kube-scheduler/app`).

This repository constitutes the **software foundation** of an MSc thesis on
cost-aware, multi-cluster scheduling for Machine Learning workloads. The current
baseline implements a minimal Filter plugin (**HelloPolitecnico**) that validates
the end-to-end out-of-tree integration on a local [`kind`](https://kind.sigs.k8s.io/)
cluster. Subsequent milestones extend this framework with Katib-aware early-stop
logic, Liqo offloading, and financial cost optimization.

| Item | Value |
|------|--------|
| Target Kubernetes | **v1.29.0** |
| Language | Go 1.22+ |
| Integration model | Out-of-tree Scheduling Framework plugins |
| Current baseline | `HelloPolitecnico` Filter plugin (proof of concept) |
| Thesis deadline | **28 February** |

---

## Background & Problem Statement

### 1. The Blindness of Default Kubernetes

The default Kubernetes scheduling framework operates efficiently for traditional microservices but is inherently "blind" to Machine Learning workloads. Its native plugins evaluate basic node metrics (e.g., CPU, Memory) but lack any contextual awareness of ML-specific lifecycle events, such as dynamic hyperparameter trial pruning (early-stopping) driven by Katib.

### 2. Architectural Flaws in Prior Art

Recent attempts to introduce AI/ML awareness into Kubernetes schedulers (e.g., DRL-MLS) have relied on anti-patterns that severely degrade performance:

- **Out-of-Tree Overheads:** Many solutions use external Python webhooks or remote SSH execution to scrape cluster state, introducing massive serialization and network latency (often >500ms per scheduling cycle).
- **Early-Stopping Hazard:** Existing reinforcement learning schedulers falsely reward algorithms when Katib prunes unpromising trials early, decoupling the reward from actual scheduling acceleration.

### 3. Our Proposed Solution

Following best practices for cloud-native systems, this thesis abandons external scripts and implements the custom logic directly as **In-Tree Go Plugins** within the native `kube-scheduler` framework. By doing so, we achieve sub-millisecond decision latency while enabling:

- **Katib Integration (PreFilter):** Safely discarding nodes for early-stopped trials and enforcing strict HPC deadline bounds.
- **Cost-Aware Federation (Filter & Score):** Dynamically triggering Liqo Virtual Kubelets to offload workloads to external clusters based on financial cost and resource availability.

---

## Thesis Roadmap & Milestones (October – February 28)

The development plan advances from a reproducible scheduler baseline to a fully
evaluated thesis contribution, culminating in monograph submission.

| Month | Milestone | Objectives |
|-------|-----------|------------|
| **October** | Baseline framework & local testbed | Establish the out-of-tree Go scheduler (`HelloPolitecnico`), containerize the binary, and validate scheduling behaviour on a local `kind` cluster (Kubernetes v1.29). |
| **November** | Cost-aware Liqo offloading & Katib integration | Implement **Financial Cost Optimization for Liqo Offloading**; integrate **Katib** metadata extraction to feed trial state, pruning signals, and deadline constraints into the scheduling cycle. |
| **December** | Benchmarking & stress evaluation | Execute comprehensive benchmarking campaigns, including **arrival-burst** stress tests; collect latency, placement, cost, and offloading metrics under controlled workloads. |
| **January** | Ablation & heterogeneous validation | Conduct ablation studies isolating plugin contributions; validate behaviour across **heterogeneous hardware** and multi-cluster topologies. |
| **February 28** | Thesis finalization | Consolidate experimental results, finalize the thesis monograph, and submit by the hard deadline of **28 February**. |

```text
Oct                Nov                 Dec                Jan               Feb 28
 |------------------|-------------------|-------------------|------------------|
 Baseline + kind     Liqo cost + Katib   Benchmarks/bursts   Ablation/hetero    Monograph
```

---

## Code Architecture Strategy

Custom logic is injected as **compiled-in Scheduling Framework plugins**,
registered from `main.go` via `app.WithPlugin(...)` and enabled through
`KubeSchedulerConfiguration`. This preserves upstream scheduler behaviour while
allowing precise control at selected extension points.

![Kubernetes Scheduling Framework Architecture](https://kubernetes.io/images/docs/scheduling-framework-extensions.png)

### Extension-point mapping

| Extension point | Intended responsibility in this thesis |
|-----------------|----------------------------------------|
| **PreFilter** | Resolve the **Katib early-stopping hazard**: use trial pruning probabilities and strict deadlines to discard infeasible scheduling candidates *before* per-node evaluation, avoiding wasted placement of trials that are statistically likely to be pruned. |
| **Filter** | Enforce hard feasibility for **Liqo multi-cluster offloading** (e.g., whether a pod may remain local or must be considered for remote execution) and apply non-negotiable cost / policy constraints. |
| **Score** | Rank feasible nodes (local vs. external cloud/peer clusters) according to **financial cost calculation** and offloading utility, preferring placements that minimise expenditure while meeting QoS and deadline objectives. |

### Design rationale

- **PreFilter** operates once per scheduling cycle and is the natural locus for
  Katib-derived global state (trial metadata, early-stop risk, deadline
  pressure). Placing this logic here prevents expensive Filter/Score work on
  pods that should not proceed.
- **Filter & Score** jointly encode the **local-vs-failover** decision for Liqo:
  Filter removes economically or topologically invalid options; Score orders the
  remaining set by monetary cost and offloading preference toward external cloud
  resources when beneficial.
- The October baseline (`HelloPolitecnico`) occupies only the **Filter** point as
  a no-op success path, proving plugin registration, configuration, and
  deployment before the November/December research logic is introduced.

---

## Project layout

```text
.
├── main.go                         # Registers plugins and starts kube-scheduler
├── pkg/plugins/hello/hello.go      # HelloPolitecnico FilterPlugin (October baseline)
├── go.mod / go.sum                 # Dependencies pinned to k8s v1.29.0
├── Dockerfile                      # Multi-stage image build
├── deploy/
│   ├── rbac.yaml                   # ServiceAccount + ClusterRoleBindings
│   ├── configmap.yaml              # KubeSchedulerConfiguration
│   └── deployment.yaml             # Scheduler Deployment
└── examples/
    └── test-pod.yaml               # Pod using schedulerName: hello-politecnico-scheduler
```

---

## How the baseline works

1. `main.go` wraps `k8s.io/kubernetes/cmd/kube-scheduler/app` and registers the
   custom plugin with `app.WithPlugin(...)`.
2. `HelloPolitecnico` implements only `FilterPlugin`.
3. On each Filter call it prints:

   ```text
   Hello Politecnico: evaluating node [<node>] for pod [<pod>]
   ```

4. Pods opt into this scheduler with:

   ```yaml
   spec:
     schedulerName: hello-politecnico-scheduler
   ```

---

## Prerequisites

- Go **1.22+**
- Docker
- [`kind`](https://kind.sigs.k8s.io/)
- `kubectl`

---

## 1. Build the binary (optional local check)

```bash
cd ~/Thesis/hello-politecnico-scheduler
go build -o hello-politecnico-scheduler .
```

---

## 2. Create a kind cluster (Kubernetes 1.29)

If you already run another kind cluster, raise inotify limits first
(common cause of create failures on Ubuntu):

```bash
sudo sysctl -w fs.inotify.max_user_watches=524288 fs.inotify.max_user_instances=512
echo -e 'fs.inotify.max_user_watches=524288\nfs.inotify.max_user_instances=512' \
  | sudo tee /etc/sysctl.d/99-kind.conf
```

Create the cluster:

```bash
kind create cluster --name hello-scheduler --image kindest/node:v1.29.0
kubectl cluster-info --context kind-hello-scheduler
```

---

## 3. Build and load the Docker image

Kind nodes do **not** see images from your host Docker daemon automatically.
You must build, then load:

```bash
docker build -t hello-politecnico-scheduler:latest .
kind load docker-image hello-politecnico-scheduler:latest --name hello-scheduler
```

---

## 4. Deploy the custom scheduler

```bash
kubectl apply -f deploy/rbac.yaml
kubectl apply -f deploy/configmap.yaml
kubectl apply -f deploy/deployment.yaml

kubectl -n kube-system rollout status deployment/hello-politecnico-scheduler
kubectl -n kube-system get pods -l component=hello-politecnico-scheduler
```

Expected: pod `Running` and `1/1 READY`.

---

## 5. Schedule a test pod

```bash
kubectl apply -f examples/test-pod.yaml
kubectl get pod hello-politecnico-test -o wide
kubectl get pod hello-politecnico-test -o jsonpath='{.spec.schedulerName}{" -> "}{.spec.nodeName}{"\n"}'
```

---

## 6. Verify the Hello Politecnico log

The Filter message is emitted **once per scheduling cycle**. `--tail=50` may miss
it if later healthz / auth noise fills the buffer. Prefer:

```bash
kubectl -n kube-system logs -l component=hello-politecnico-scheduler --tail=500 \
  | grep "Hello Politecnico"
```

Expected output:

```text
Hello Politecnico: evaluating node [hello-scheduler-control-plane] for pod [hello-politecnico-test]
```

To force a fresh Filter call:

```bash
kubectl delete pod hello-politecnico-test --force --grace-period=0
kubectl apply -f examples/test-pod.yaml
kubectl -n kube-system logs -l component=hello-politecnico-scheduler --tail=30 \
  | grep "Hello Politecnico"
```

---

## Live demo (for your professor)

Run these four steps with the cluster and scheduler already deployed:

```bash
# 1) Show the custom scheduler is running
kubectl -n kube-system get pods -l component=hello-politecnico-scheduler

# 2) Delete the old test pod and create a fresh one (triggers Filter again)
kubectl delete pod hello-politecnico-test --force --grace-period=0
kubectl apply -f examples/test-pod.yaml

# 3) Show it was scheduled by YOUR scheduler
kubectl get pod hello-politecnico-test -o wide
kubectl get pod hello-politecnico-test -o jsonpath='Scheduler={.spec.schedulerName}  Node={.spec.nodeName}{"\n"}'

# 4) Show the Hello Politecnico log
kubectl -n kube-system logs -l component=hello-politecnico-scheduler --tail=50 \
  | grep "Hello Politecnico"
```

Expected log line:

```text
Hello Politecnico: evaluating node [hello-scheduler-control-plane] for pod [hello-politecnico-test]
```

**Talking points while demoing:**

1. The pod sets `schedulerName: hello-politecnico-scheduler` (not the default scheduler).
2. The Go `HelloPolitecnico` Filter plugin runs during the scheduling cycle.
3. The log line proves the plugin evaluated the node and returned `Success`.

Optional live follow (two terminals):

```bash
# Terminal A
kubectl -n kube-system logs -l component=hello-politecnico-scheduler -f \
  | grep --line-buffered "Hello Politecnico"

# Terminal B
kubectl delete pod hello-politecnico-test --force --grace-period=0
kubectl apply -f examples/test-pod.yaml
```

---

## Troubleshooting

| Symptom | Fix |
|--------|-----|
| `ImagePullBackOff` / cannot pull image | Run `kind load docker-image hello-politecnico-scheduler:latest --name hello-scheduler`, then delete the scheduler pod so it restarts |
| kind create fails with `Reached target .*Multi-User System` | Raise inotify limits (see above) and/or delete other kind clusters |
| `grep "Hello Politecnico"` prints nothing | Increase `--tail` or recreate the test pod; the message appears only when a pod is being scheduled |
| ConfigMap `extension-apiserver-authentication` forbidden warnings | Harmless for this demo (auth webhook probing). Scheduling still works |

---

## Cleanup

```bash
kubectl delete -f examples/test-pod.yaml
kubectl delete -f deploy/
kind delete cluster --name hello-scheduler
```

---

## Notes

- The October baseline is intentionally **minimal** (Filter only) to prove
  out-of-tree compilation, configuration, and deployment.
- Default scheduler plugins remain enabled; `HelloPolitecnico` is added to the
  Filter chain via `KubeSchedulerConfiguration`.
- Leader election is disabled (`leaderElect: false`) for a single-replica demo.
- `go.mod` includes the required `replace` directives so
  `k8s.io/kubernetes@v1.29.0` compiles as an out-of-tree dependency.
- PreFilter / cost-aware Score plugins described above are **planned** research
  extensions and are not yet present in this baseline snapshot.
