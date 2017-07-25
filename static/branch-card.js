// custom directive to keep a relative time up to date
Vue.directive('moment-ago', {
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

Vue.component('branch-card', {
	data() {
		return {
			showDetail: false
		}
	},
    props: ['branch'],
	filters: {
		truncate(text, length) {
			return text.slice(0, length);
		}
	},
	template: `
        <div class="card text-center" v-bind:class="{ 'card-outline-danger': branch.state === 'failure', 'card-outline-success': branch.state === 'success' }">
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
                        &middot; <small class="text-muted"><span v-moment-ago="branch.last_updated"></span></small>
                    </span>
                    
                    &middot;
                    
                    <small class="code text-muted">
                        <a v-bind:href="branch.commit_url" class="deco-none">{{ branch.sha | truncate(7) }}</a>
                    </small>
                </p>
                
            </div>
            <div class="list-group list-group-flush text-left">
                <template v-for="status in branch.status_checks">
                    <a v-if="status.status_url" v-bind:href="status.status_url" class="list-group-item list-group-item-action deco-none">
                        <small>
                            <span v-if="status.state === 'failure'">❌</span>                
                            {{ status.description }}
                        </small>
                    </a>
                    <div v-else class="list-group-item">
                        <small>
                            <span v-if="status.state === 'failure'">❌</span>                
                            {{ status.description }}
                        </small>
                    </div>
                </template>
            </div>
        </div>
	`
});
