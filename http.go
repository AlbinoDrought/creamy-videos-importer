package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/AlbinoDrought/creamy-videos-importer/creamqueue"
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

		form {
			display: flex;
			flex-direction: row;
		}
		input {
			flex: 1;
			margin: 0 1em;
			outline: none;
			width: 100%;
			border: 1px solid rgba(34, 36, 38, 0.15);
			background-color: rgba(255, 255, 255, 0.1);
			color: white;
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
		<form method="POST">
			<label for="url">URL</label>
			<input type="text" name="url">

			<button type="submit">Queue</button>
		</form>
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
		</table>

		<script type="text/javascript">
			if (fetch) {
				var fetching = false;
				var localTable = document.querySelector('table');

				setInterval(function () {
					if (fetching) {
						return;
					}

					fetch('/?autofetch').then(function (resp) {
						return resp.text();
					}).then(function (text) {
						var el = document.createElement('html');
						el.innerHTML = text;

						var remoteTable = el.querySelector('table');
						if (localTable && remoteTable) {
							localTable.innerHTML = remoteTable.innerHTML;
						}

						fetching = false;
					}).catch(function (ex) {
						console.error('error fetching', ex);
						fetching = false;
					})
				}, 5000);
			}
		</script>
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

func handlerCreateJob(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(400)
		w.Write([]byte("bad data"))
		return
	}

	url := r.FormValue("url")
	if url == "" {
		w.WriteHeader(422)
		w.Write([]byte("missing \"url\" value"))
		return
	}

	queue.Push(idGenerator.Next(), creamqueue.JobData{
		URL: url,
	})

	http.Redirect(w, r, "/", 302)
}

func bootServer(ctx context.Context) chan error {
	router := makeRouter([]routeDef{
		routeDef{"GET", "/", "ViewJobs", handlerViewJobs},
		routeDef{"POST", "/", "CreateJob", handlerCreateJob},
	})

	src := &http.Server{
		Addr:    ":" + config.port,
		Handler: router,
	}

	errorChannel := make(chan error, 1)

	go func() {
		log.Println("listening on", src.Addr)
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
