# Release v0.0.1

<br/>

---
## Notes

Thank you to all contributors who helped shape the initial release of KubeMedic! This first version lays the foundation for automated Kubernetes remediation. Please try it out and let us know if you encounter any issues.

<br/>

---
## Change Logs

+ Initial release of KubeMedic operator for automated Kubernetes remediation
+ Added SelfRemediationPolicy Custom Resource Definition (CRD) for defining remediation rules and conditions
+ Implemented core monitoring capabilities:
    ```
    - CPU utilization
    - Memory usage 
    - Error rates
    - Pod restart counts
    ```
+ Added automated remediation actions:
    - Scale up deployments
    - Scale down deployments 
    - Restart problematic pods
    - Rollback deployments to last known good state
+ Integrated with Grafana for enhanced visualization (optional)
+ Added basic Prometheus metrics for operator monitoring
+ Included example configurations and use cases in /examples directory
+ Implemented RBAC security controls:
    - Protected system namespaces
    - Resource-level protections
    - Configurable restrictions

<br/>

---
## Known Issues

+ Initial release - no known major issues
+ Please report any bugs or feature requests through GitHub issues