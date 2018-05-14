// +build !integration

package node

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	mbtest "github.com/elastic/beats/metricbeat/mb/testing"
)

func TestFetch(t *testing.T) {

	files, err := filepath.Glob("./_meta/test/node.*.json")
	assert.NoError(t, err)
	// Makes sure glob matches at least 1 file
	assert.True(t, len(files) > 0)

	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			response, err := ioutil.ReadFile(f)
			assert.NoError(t, err)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Header().Set("Content-Type", "application/json;")
				w.Write([]byte(response))
				assert.Equal(t, "/_nodes/_local", r.RequestURI)
			}))
			defer server.Close()

			config := map[string]interface{}{
				"module":     "elasticsearch",
				"metricsets": []string{"node"},
				"hosts":      []string{server.URL},
			}
			reporter := &mbtest.CapturingReporterV2{}

			metricSet := mbtest.NewReportingMetricSetV2(t, config)
			metricSet.Fetch(reporter)

			e := mbtest.StandardizeEvent(metricSet, reporter.GetEvents()[0])
			t.Logf("%s/%s event: %+v", metricSet.Module().Name(), metricSet.Name(), e.Fields.StringToPrint())
		})
	}
}
