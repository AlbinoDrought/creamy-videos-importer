package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"
)

const rawTemplateViewJobs = `
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>Creamy Videos Importer</title>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<style type="text/css">
		html, body {
			font-family: mono;
			background-color: #1b1b1b;
			color: #ababab;
		}
		a, a:visited {
			color: mediumaquamarine;
		}

		table {
			width: 100%;
		}
		th {
			text-align: left;
		}
		th, td {
			padding: 0.5em;
		}
		td {
			border-top: 1px solid rgba(255,255,255,0.2);
			border-collapse: collapse;
		}

		.status--finished { color: lawngreen; }
		.status--failed { color: crimson; }
		.status--started { color: cornflowerblue; }
		</style>
	</head>
	<body>
		<table>
			<thead>
				<tr>
					<th>Job</th>
					<th>Queued At</th>
					<th>Runtime</th>
					<th>Status</th>
				</tr>
			</thead>
			<tbody>
				{{ range $element := .Jobs }}
					<tr>
						<td>
							<strong>Input:</strong>
							<a href="{{ $element.Data.URL }}">
								{{ $element.Data.URL }}
							</a>

							{{ if (eq $element.Status "finished") }}
								<br>
								{{ if (eq $element.Result.CreamyURL "" ) }}
									{{ $element.Result.Title }}
								{{ else }}
									<strong>Video:</strong>
									<a href="{{ $element.Result.CreamyURL }}">
										{{ $element.Result.Title }}
									</a>
								{{ end }}
							{{ end }}
							{{ if (eq $element.Status "failed") }}
								<br>
								<strong>Failure Reasons:</strong>
								<ul>
									{{ range $failure := $element.Failures }}
										<li>{{ $failure.Error }}</li>
									{{ end }}
								</ul>
							{{ end }}
						</td>
						<td>{{ humanTime $element.CreatedAt }}</td>
						<td>{{ runtime $element }}</td>
						<td class="status status--{{ $element.Status }}">
							{{ $element.Status }}
						</td>
					</tr>
				{{ end }}
			</tbody>
	</body>
</html>
`

var templateViewJobs = template.Must(template.New("viewJobs").Funcs(template.FuncMap{
	"humanTime": func(timestamp time.Time) string {
		if timestamp.IsZero() {
			return ""
		}
		return timestamp.Format(time.Stamp)
	},
	"runtime": func(job *jobInformation) string {
		if job.StartedAt.IsZero() || job.StoppedAt.IsZero() {
			return "-"
		}

		return job.StoppedAt.Sub(job.StartedAt).Truncate(time.Millisecond).String()
	},
}).Parse(rawTemplateViewJobs))

func handlerViewJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")

	jobRepo.lock.RLock()
	defer jobRepo.lock.RUnlock()

	jobs := make([]*jobInformation, len(jobRepo.jobs))

	i := 0
	for _, job := range jobRepo.jobs {
		job.lock.RLock()
		jobs[i] = job
		defer job.lock.RUnlock()
		i++
	}

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt.After(jobs[j].CreatedAt)
	})

	err := templateViewJobs.Execute(w, struct {
		Jobs []*jobInformation
	}{jobs})

	if err != nil {
		log.Println("error rendering viewJobs template:", err)
	}
}

func bootServer(ctx context.Context) chan error {
	router := makeRouter([]routeDef{
		routeDef{"GET", "/", "ViewJobs", handlerViewJobs},
	})

	src := &http.Server{
		Addr:    ":4000",
		Handler: router,
	}

	errorChannel := make(chan error, 1)

	go func() {
		errorChannel <- src.ListenAndServe()
	}()

	go func() {
		select {
		case <-ctx.Done():
			gracefulCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()
			src.Shutdown(gracefulCtx)
			<-errorChannel
		}
	}()

	return errorChannel
}
