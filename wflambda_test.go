package wflambda

import (
	"os"
	"testing"
)

func TestGetAndValidateLambdaEnvironment(t *testing.T) {
	os.Setenv("WAVEFRONT_URL", "https://demo.wavefront.com")
	os.Setenv("WAVEFRONT_API_TOKEN", "demo-api-token")
	os.Setenv("REPORT_STANDARD_METRICS", "False")
	expected := getAndValidateLambdaEnvironment()
	if expected != false {
		t.Error("Validate environmental variables failed ", expected, "False")
	}
	os.Setenv("REPORT_STANDARD_METRICS", "true")
	expected = getAndValidateLambdaEnvironment()
	if expected != true {
		t.Error("Validate environmental variables failed ", expected, "true")
	}
}
