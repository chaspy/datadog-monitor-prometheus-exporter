package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	//nolint:gochecknoglobals
	DatadogMonitorAlert = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "datadog_monitor",
		Subsystem: "prometheus_exporter",
		Name:      "alert_count",
		Help:      "The number of alert count",
	},
		[]string{"name", "priority", "tags"},
	)
)

func main() {
	if err := core(); err != nil {
		log.Fatal(err)
	}
}

func core() error {
	interval, err := getInterval()
	if err != nil {
		return err
	}

	err = readDatadogConfig()
	if err != nil {
		return fmt.Errorf("failed to read Datadog Config: %w", err)
	}

	prometheus.MustRegister(DatadogMonitorAlert)

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)

		// register metrics as background
		for range ticker.C {
			err := snapshot()
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return fmt.Errorf("failed to listen HTTP server: %w", err)
	}
	return nil
}

func snapshot() error {
	DatadogMonitorAlert.Reset()

	monitors, err := getMonitors()
	if err != nil {
		return fmt.Errorf("failed to get monitors: %w", err)
	}

	// monitorInfos := getMonitorInfos(monitors)

	/*
		for _, monitorInfo := range monitorInfos {
			labelsTag := make([]string, len(monitorInfo.Labels))

			for i, label := range issueInfo.Labels {
				labelsTag[i] = *label.Name
			}

			labels := prometheus.Labels{
				"name":     strconv.Itoa(monitorInfo.Number),
				"label":    strings.Join(labelsTag, ","),
				"priority": monitorInfo.User,
			}
			DatadogMonitorAlert.With(labels).Set(1)
		}

	*/

	fmt.Printf("%v\n", monitors)
	return nil
}

func getInterval() (int, error) {
	const defaultAPIIntervalSecond = 300
	apiInterval := os.Getenv("API_INTERVAL")
	if len(apiInterval) == 0 {
		return defaultAPIIntervalSecond, nil
	}

	integerAPIInterval, err := strconv.Atoi(apiInterval)
	if err != nil {
		return 0, fmt.Errorf("failed to read API_INTERVAL: %w", err)
	}

	return integerAPIInterval, nil
}

func readDatadogConfig() error {
	ddAPIKey := os.Getenv("DD_API_KEY")
	if len(ddAPIKey) == 0 {
		return fmt.Errorf("missing environment variable: DD_API_KEY")
	}

	ddAPPKey := os.Getenv("DD_APP_KEY")
	if len(ddAPPKey) == 0 {
		return fmt.Errorf("missing environment variable: DD_APP_KEY")
	}

	return nil
}

func getMonitors() ([]byte, error) {
	ctx := datadog.NewDefaultContext(context.Background())

	groupStates := ""      // string | When specified, shows additional information about the group states. Choose one or more from `all`, `alert`, `warn`, and `no data`. (optional)
	name := ""             // string | A string to filter monitors by name. (optional)
	tags := ""             // string | A comma separated list indicating what tags, if any, should be used to filter the list of monitors by scope. For example, `host:host0`. (optional)
	monitorTags := ""      // string | A comma separated list indicating what service and/or custom tags, if any, should be used to filter the list of monitors. Tags created in the Datadog UI automatically have the service key prepended. For example, `service:my-app`. (optional)
	withDowntimes := false // bool | If this argument is set to true, then the returned data includes all current downtimes for each monitor. (optional)
	idOffset := int64(0)   // int64 | Monitor ID offset. (optional)
	page := int64(0)       // int64 | The page to start paginating from. If this argument is not specified, the request returns all monitors without pagination. (optional)
	pageSize := int32(100) // int32 | The number of monitors to return per page. If the page argument is not specified, the default behavior returns all monitors without a `page_size` limit. However, if page is specified and `page_size` is not, the argument defaults to 100. (optional)
	optionalParams := datadog.ListMonitorsOptionalParameters{
		GroupStates:   &groupStates,
		Name:          &name,
		Tags:          &tags,
		MonitorTags:   &monitorTags,
		WithDowntimes: &withDowntimes,
		IdOffset:      &idOffset,
		Page:          &page,
		PageSize:      &pageSize,
	}

	configuration := datadog.NewConfiguration()

	apiClient := datadog.NewAPIClient(configuration)

	resp, r, err := apiClient.MonitorsApi.ListMonitors(ctx, optionalParams)

	// debug
	// fmt.Printf("%v", r)
	// debug
	// fmt.Printf("%v", err)
	// debug
	// fmt.Printf("%v", resp)

	if err != nil {
		return []byte(""), fmt.Errorf("failed to call `MonitorsApi.ListMonitors`: %w", err)
	}
	defer r.Body.Close()

	// response from `ListMonitors`: []Monitor
	responseContent, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return []byte(""), fmt.Errorf("failed to marshal json: %w", err)
	}

	// debug
	fmt.Fprintf(os.Stdout, "Response from MonitorsApi.ListMonitors:\n%s\n", responseContent)

	return responseContent, nil
}

/*
func getDatadogMonitorInfos(issues []*github.Issue) []Issue {
	issueInfos := make([]Issue, len(issues))

	for i, issue := range issues {
		repos := strings.Split(*issue.URL, "/")

		issueInfos[i] = Issue{
			Number: issue.GetNumber(),
			Labels: issue.Labels,
			User:   issue.User.GetLogin(),
			Repo:   repos[4] + "/" + repos[5],
		}
	}

	return issueInfos
}
*/
