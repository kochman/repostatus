// handle websocket connection to backend
function createWebSocket(onmessage, onopen) {
	// determine relative path to /ws endpoint
	var loc = window.location, wsURI;
	if (loc.protocol === "https:") {
		wsURI = "wss:";
	} else {
		wsURI = "ws:";
	}
	wsURI += "//" + loc.host + "/ws";
	var ws = new WebSocket(wsURI);

	ws.onmessage = onmessage;
	ws.onopen = onopen;
	ws.onclose = function(event) {
		console.log("closed, restarting");
		setTimeout(function() {
			createWebSocket(onmessage, onopen);
		}, 1000);
	}
}

// custom directive to keep a relative time up to date
app.directive('moment-ago', {
	inserted(el, binding) {
		const timestamp = binding.value;
		el.innerHTML = moment(timestamp).fromNow();

		this.interval = setInterval(() => {
			el.innerHTML = moment(timestamp).fromNow();
		}, 1000)
	},
	unbind() {
		clearInterval(this.interval);
	}
});

app.component('repo', {
	data() {
		return {
			branches: [],
			repo: {},
			loaded: false
		}
	},
	created() {
		var vue = this;
		var onmessage = function(event) {
		    var repo = JSON.parse(event.data);
		    vue.repo = repo;
			vue.loaded = true;
		}
		var onopen = function(event) {
			var ws = event.target;
			var org = vue.$route.params.org;
			var repo = vue.$route.params.repo;
			var msg = {command: "subscribe", data: {org: org, repo: repo}};
			ws.send(JSON.stringify(msg));
		}
		createWebSocket(onmessage, onopen);
	},
	template: `
		<div v-if="loaded">
            <div class="container">
            	<div class="row mt-4">
            		<div class="col-12 text-center">
                        <p class="h1">{{ repo.name }}</p>
                        <p>
                        	<a v-bind:href="repo.github_url">GitHub</a>
                            <template v-if="repo.website"> &middot; <a v-bind:href="repo.website">Website</a> </template>
                        </p>
                        <p class="text-secondary">{{ repo.description }}</p>
                    </div>
                </div>
            </div>
            <div class="branches">
                <div class="container-fluid card-columns">
                    <branch-card v-bind:branch="branch" class="mb-3" v-for="branch in repo.branches" v-bind:key="branch.name"></branch-card>
                </div>
            </div>
        </div>
        <div v-else>
            <div class="spinner" v-if="!loaded"></div>
        </div>
	`
});
