# VMClarity UI Tour
Figure 1. VMClarity UI Dashboard
![VMClarity UI Dashboard](img/vmclarity-ui-1.png)

## Configure Your First Scan

- Click on the "Scans" icon as shown in Figure 2. In the Scans window, you can create a new scan configuration.

Figure 2. VMClarity UI Scan

<img src="img/vmclarity-ui-2.png" alt="VMClarity UI Scan" width="25%" height="25%" title="VMClarity UI Scan" />

- Create a new scan configuration. As shown in Figure 3, click on "New scan configuration".

Figure 3. VMClarity Scan Setup Step 1

<img src="img/vmclarity-scan-setup-1.png" alt="VMClarity Scan Setup - Step 1" width="40%" height="40%" title="VMClarity Scan Setup Step 1" />

- In the "New scan config" wizard shown in Figure 4, follow the wizard steps to name the scan, and identify the AWS scope (region, VPC, security groups, etc). In the example shown in Figure 4, the AWS us-east-2 region, and a specific VPC were identied as well as a specific EC2 instance with the name "vmclarity-demo-vm".

Figure 4. VMClarity Scan Setup Step 2

<img src="img/vmclarity-scan-setup-2.png" alt="VMClarity Scan Setup - Step 2" width="40%" height="40%" title="VMClarity Scan Setup Step 2" />

- Next, identify all of the scan types you want enabled. As Figure 5 shows, all of the available scan types have been enabled.

Figure 5. VMClarity Scan Setup Step 3

<img src="img/vmclarity-scan-setup-3.png" alt="VMClarity Scan Setup - Step 3" width="40%" height="40%" title="VMClarity Scan Setup Step 3" />

- Finally, select the scan time and/or frequency of the scans. Figure 6 shows the scan option of "now", but other options include "specific time" and "recurring" (Based on a cron job).

Figure 6. VMClarity Scan Setup Step 4

<img src="img/vmclarity-scan-setup-4.png" alt="VMClarity Scan Setup - Step 4" width="40%" height="40%" title="VMClarity Scan Setup Step 4" />

- Once all of the scan setup steps have been entered, click on "Save".

In the Scan Configurations tab, you will see the scan config listed as shown in Figure 7.

Figure 7. VMClarity Scan Configuration List

<img src="img/vmclarity-scan-config-summary.png" alt="VMClarity Scan Config Summary" width="90%" height="90%" title="VMClarity Scan Config Summary" />

Once a scan runs and generates findings, you can browse around the various VMClarity UI features and investigate the security scan reports.

Here are a few of the many scan findings that are available in the VMClarity UI.

Figure 8. VMClarity Scan List

<img src="img/vmclarity-scan-list.png" alt="VMClarity Scan List" width="90%" height="90%" title="VMClarity Scan List" />

Figure 9. VMClarity Dashboard

<img src="img/vmclarity-dashboard-data.png" alt="VMClarity Dashboard with Findings" width="90%" height="90%" title="VMClarity Dashboard with Findings" />
