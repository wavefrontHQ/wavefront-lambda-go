package wflambda

import (
	"os"
	"testing"
)

func TestGetStandardLambdaMetricName(t *testing.T) {
	actualName := getStandardLambdaMetricName("customMetric", true)
	expectedName := "aws.lambda.wf.customMetric_event"
	if actualName != expectedName {
		t.Error("Metric names don't match ", expectedName, actualName)
	}
	actualName = getStandardLambdaMetricName("customMetrics", false)
	expectedName = "aws.lambda.wf.customMetrics"
	if actualName != expectedName {
		t.Error("Metric names don't match ", expectedName, actualName)
	}
}

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
