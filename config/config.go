package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SitemapUrl               string        `yaml:"sitemap-url"`
	ReportName               string        `yaml:"report-name"`
	AggregateReportName      string        `yaml:"aggregate-report-name"`
	LogName                  string        `yaml:"log-name"`
	UserAgent                string        `yaml:"user-agent"`
	FilterInputUrl           string        `yaml:"filter-input-url"`
	FilterOutputStatus       string        `yaml:"filter-output-status"`
	ConcurrentRequests       int           `yaml:"concurrent-requests"`
	ConnectionTimeout        int           `yaml:"connection-timeout"`
	Limit                    int           `yaml:"limit"`
	SkipLogs                 bool          `yaml:"skip-logs"`
	SkipReport               bool          `yaml:"skip-report"`
	SkipAggregateReport      bool          `yaml:"skip-aggregate-report"`
	UseDateSuffix            bool          `yaml:"use-date-suffix"`
	DryRun                   bool          `yaml:"dry-run"`
	FilterOutputResponseTime time.Duration `yaml:"filter-output-response-time"`
}

var DefaultConfig = &Config{
	SitemapUrl:               "",
	ReportName:               "cachewarmup_report.csv",
	AggregateReportName:      "cachewarmup_aggregate_report.csv",
	LogName:                  "cachewarmup.log",
	UserAgent:                "GoCacheWarmer",
	FilterInputUrl:           "",
	FilterOutputStatus:       "",
	ConcurrentRequests:       10,
	ConnectionTimeout:        30,
	Limit:                    0,
	SkipLogs:                 false,
	SkipReport:               false,
	SkipAggregateReport:      false,
	UseDateSuffix:            false,
	DryRun:                   false,
	FilterOutputResponseTime: time.Duration(0),
}

func InitCliConfig() *Config {
	config := *DefaultConfig
	if data, err := os.ReadFile("config.yaml"); err == nil {
		yaml.Unmarshal(data, &config)
	}

	flag.StringVar(&config.SitemapUrl, "sitemap-url", config.SitemapUrl, "Sitemap URL")
	flag.StringVar(&config.ReportName, "report-name", config.ReportName, "Report file name")
	flag.StringVar(&config.AggregateReportName, "aggregate-report-name", config.AggregateReportName, "Aggregate report file name")
	flag.StringVar(&config.LogName, "log-name", config.LogName, "Log file name")
	flag.StringVar(&config.UserAgent, "user-agent", config.UserAgent, "User agent")
	flag.StringVar(&config.FilterInputUrl, "filter-input-url", config.FilterInputUrl, "Filter URLs from sitemap that will be warmed up, allowed * symbol for pattern matching (for example: 2XX, 500)")
	flag.StringVar(&config.FilterOutputStatus, "filter-output-status", config.FilterOutputStatus, "Filter URLs in report by status mask")

	flag.IntVar(&config.ConcurrentRequests, "concurrent-requests", config.ConcurrentRequests, "Amount of concurrent HTTP-requests")
	flag.IntVar(&config.ConnectionTimeout, "connection-timeout", config.ConnectionTimeout, "Connection timeout in seconds for HTTP-requests")
	flag.IntVar(&config.Limit, "limit", config.Limit, "Amount of URLs from sitemap that will be processed")

	flag.BoolVar(&config.SkipLogs, "skip-logs", config.SkipLogs, "Skip logs")
	flag.BoolVar(&config.SkipReport, "skip-report", config.SkipReport, "Skip main report creation")
	flag.BoolVar(&config.SkipAggregateReport, "skip-aggregate-report", config.SkipAggregateReport, "Skip aggregate report creation")

	flag.BoolVar(&config.DryRun, "dry-run", config.DryRun, "Will imitate warmup requests instead of doing real HTTP-requests")

	flag.DurationVar(&config.FilterOutputResponseTime, "filter-output-response-time", config.FilterOutputResponseTime, "Filter URLs in report above threshold (for example: 200ms, 1.5s)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "run: execute cache warmup\n")
		fmt.Fprintf(os.Stderr, "generate-config: generate YAML-config file\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	return &config
}
