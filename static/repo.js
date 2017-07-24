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
Vue.directive('moment-ago', {
	inserted (el, binding) {
		const timestamp = binding.value;
		el.innerHTML = moment(timestamp).fromNow();

		this.interval = setInterval(() => {
			el.innerHTML = moment(timestamp).fromNow();
		}, 1000)
	},
	unbind () {
		clearInterval(this.interval);
	}
});

Vue.component('repo', {
	data() {
		return {
			branches: [],
			loaded: false
		}
	},
	created: function() {
		var vue = this;
		var onmessage = function(event) {
			var branches = JSON.parse(event.data);
			while (vue.branches.length > 0) {
				vue.branches.pop();
			}
			for (var branch of branches) {
				vue.branches.push(branch);
			}
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
	filters: {
		relativeTime(time) {
			return moment(time).fromNow();
		}
	},
	template: `
		<div class="branches" v-if="loaded">
			<div class="container-fluid card-columns">
				<div class="card text-center mb-3" v-for="branch in branches" v-if="branch.state" v-bind:class="{ 'card-outline-danger': branch.state === 'failure', 'card-outline-success': branch.state === 'success' }">
					<div class="card-block">
						<h4 class="card-title">
							<a v-bind:href="branch.commits_url" class="deco-none">{{ branch.name }}</a>
						</h4>

						<p class="card-text">
							<span class="text-success" v-if="branch.state === 'success'">
								Success
							</span>
							<span class="text-danger" v-if="branch.state === 'failure'">
								Failure
							</span>
							<span class="" v-if="branch.state === 'pending'">
								No status checks
							</span>
							<span v-if="branch.state === 'success' || branch.state === 'failure'" class="">
								<small class="text-muted"><span v-moment-ago="branch.last_updated"></span></small>
							</span>
						</p>

					</div>
				</div>
			</div>
		</div>
		<div class="spinner" v-else></div>
	`
});
