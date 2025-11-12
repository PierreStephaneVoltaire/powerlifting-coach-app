# Claude Development Instructions

## Code Quality Guidelines

### Minimize Fallback Logic
- Avoid excessive fallback chains and nested conditionals
- Prefer explicit error handling over silent fallbacks
- When fallbacks are necessary, keep them simple and direct

### Code Comments
- **Only include non-obvious "why" explanations**
- **Do NOT include "what" comments that describe what the code does**
- Comment on business logic rationale, architectural decisions, and workarounds
- Let the code be self-documenting for obvious operations

## Kubernetes Operations

### Authentication
When running kubectl commands, always use the kubeconfig file:
```bash
export KUBECONFIG=/mnt/c/Users/Pierre/Documents/powerlifting-coach-app/infrastructure/kubeconfig.yaml
```

## Pre-Commit Validation

### Frontend and Service Changes
When modifying files in `frontend/` or `services/`, **MUST** run docker build before committing:
```bash
# For frontend changes
docker build -t <image-name>:<tag> ./frontend

# For service changes
docker build -t <image-name>:<tag> ./services/<service-name>
```

### Infrastructure Changes
When modifying files in `infrastructure/`, **MUST** run validation and formatting before committing:
```bash
cd infrastructure
terraform fmt -recursive
terraform validate
```

## Resource Provisioning

### Cost Optimization
- **Always provision the cheapest viable resources** for Kubernetes and Terraform
- Use minimal node sizes, instance types, and resource allocations
- Only provision more expensive resources when explicitly instructed by the user
- Examples:
  - Use smallest viable node pools
  - Prefer spot/preemptible instances when appropriate
  - Minimize replica counts
  - Use resource limits conservatively

### Exception Handling
Only deviate from cheap resource provisioning when the user:
- Explicitly requests higher performance
- Specifies production-grade resources
- Provides specific resource requirements

## Debugging Tools

### Installation Policy
When debugging tools are needed but not installed:
1. **List all required tools** with their purpose
2. **Ask the user** to either:
   - Install the tools manually
   - Provide alternative instructions
3. **Do NOT** automatically install tools without permission
4. **Wait for user confirmation** before proceeding

Example response:
```
The following debugging tools are required:
- `kubectl-debug`: For ephemeral container debugging
- `stern`: For multi-pod log streaming
- `k9s`: For interactive cluster exploration

Please install these tools or provide alternative instructions.
```

## Summary Checklist

- [ ] Minimize fallback logic
- [ ] Only "why" comments, no "what" comments
- [ ] Use infrastructure/kubeconfig.yaml for kubectl
- [ ] Docker build before committing frontend/services
- [ ] Terraform fmt + validate before committing infrastructure
- [ ] Default to cheap resources for k8s/terraform
- [ ] List and ask before installing debugging tools
