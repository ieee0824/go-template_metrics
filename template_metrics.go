package template_metrics

import (
	"fmt"
	metrics "github.com/yvasiyarov/go-metrics" // max,mean,min,stddev,percentile
	"html/template"                            // ToDo: how can I change this from outside?
	"io"
	"time"
)

// print infomation on each request
var Verbose = false

// a set of proxies
var proxyRegistry = make(map[string](*proxyTemplate))

type proxyTemplate struct {
	name     string
	original *template.Template
	timer    metrics.Timer
}

func newProxyTemplate(name string, template *template.Template) *proxyTemplate {
	return &proxyTemplate{
		name:     name,
		original: template,
		timer:    metrics.NewTimer(),
	}
}

//print the elapsed time on each request if Verbose flag is true
func (proxy *proxyTemplate) printVerbose(elapsedTime time.Duration) {
	fmt.Printf("time:%v\ttemplate:%s\telapsed:%f\n",
		time.Now(),
		proxy.original.Name(),
		elapsedTime.Seconds(),
	)
}

//measure the time
func (proxy *proxyTemplate) measure(startTime time.Time) {
	elapsedTime := time.Now().Sub(startTime)
	proxy.timer.Update(elapsedTime)
	if Verbose {
		proxy.printVerbose(elapsedTime)
	}
}

////// instrument functions

// instrument template.Execute
func (proxy *proxyTemplate) Execute(wr io.Writer, data interface{}) error {
	startTime := time.Now()
	error := proxy.original.Execute(wr, data)
	defer proxy.measure(startTime)
	return error
}

// instrucment template.ExecuteTemplate
func (proxy *proxyTemplate) ExecuteTemplate(wr io.Writer, name string, data interface{}) error {
	startTime := time.Now()
	error := proxy.original.ExecuteTemplate(wr, name, data)
	defer proxy.measure(startTime)
	return error
}

///// package functions

//Wrap  instrument template
func WrapTemplate(name string, template *template.Template) *proxyTemplate {
	proxy := newProxyTemplate(name, template)
	proxyRegistry[template.Name()] = proxy
	return proxy
}

//Print  print the metrics in each second
func Print() {
	go func() {
		time.Sleep(1000 * time.Millisecond)
		for {
			startTime := time.Now()
			for name, proxy := range proxyRegistry {
				timer := proxy.timer
				count := timer.Count()
				if count > 0 {
					fmt.Printf("time:%v\ttemplate:%s\tcount:%d\tmax:%f\tmean:%f\tmin:%f\tpercentile95:%f\n",
						time.Now(),
						name,
						timer.Count(),
						float64(timer.Max())/float64(time.Second),
						timer.Mean()/float64(time.Second),
						float64(timer.Min())/float64(time.Second),
						timer.Percentile(0.95)/float64(time.Second),
					)
					proxy.timer = metrics.NewTimer()
				}
			}
			elapsedTime := time.Now().Sub(startTime)
			time.Sleep(1000*time.Millisecond - elapsedTime)
		}
	}()
}