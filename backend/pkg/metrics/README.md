# Prometheus

In order to expose metrics from KubeClarity to Prometheus, you need to enable
the endpoint by setting the environment variable `PROMETHEUS_REFRESH_INTERVAL_SECONDS`
to value larger than zero. There is no particular need to scrape these metrics very
often, so once every 300 seconds is fine.

You can set this value in the `values.yaml` file.

The metrics that are exposed are:

| Metric                                              | Explanation                                        |
|-----------------------------------------------------|----------------------------------------------------|
| kubeclarity_application_vulnerability               | Count of vulnerabilities per application and level |
| kubeclarity_number_of_applications                  | The total number of applications                   |
| kubeclarity_number_of_fixable_vulnerabilities       | The number of fixable vulnerabilities per severity |
| kubeclarity_number_of_fixable_vulnerabilities_total | The total number of fixable vulnerabilities        |
| kubeclarity_number_of_packages                      | The total number of packages                       |
| kubeclarity_number_of_resources                     | The total number of resources                      |
| kubeclarity_number_of_vulnerabilities               | The number of vulnerabilities per severity         |
| kubeclarity_number_of_vulnerabilities_total         | The total number of vulnerabilities                |

## Prometheus alert

For an example of how to get a Prometheus alert when new issues in the cluster are found, see:
[alertmanager-kubeclarity.yaml](alertmanager-kubeclarity.yaml)

## Grafana

The file [kubeclarity-dashboard.json](kubeclarity-dashboard.json) contains a dashboard which
you can add to your Grafana instance.
