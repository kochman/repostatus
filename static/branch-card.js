function momentUpdater(el, binding) {
    const timestamp = binding.value;
    el.innerHTML = moment(timestamp).fromNow();

    clearInterval(el.interval);
    el.interval = setInterval(() => {
        el.innerHTML = moment(timestamp).fromNow();
    }, 1000);
}

// custom directive to keep a relative time up to date
app.directive('moment-ago', {
	mounted(el, binding) {
	    momentUpdater(el, binding);
	},
    updated(el, binding) {
	    momentUpdater(el, binding);
	},
	unmounted() {
		clearInterval(this.interval);
	},
});

app.component('branch-card', {
	data() {
		return {
			showDetail: false
		}
	},
    props: ['branch'],
    computed: {
		branchState: function() {
			if (this.branch.state === 'success') {
				return 'success';
			} else if (this.branch.state === 'failure') {
				return 'failure';
			} else if (this.branch.status_checks === null || this.branch.status_checks.length === 0) {
				return 'no checks';
			} else {
				return 'pending';
			}
		},
        truncatedSha: function() {
            return this.branch.sha.slice(0, 7);
        },
	},
	template: `
        <div class="card text-center" v-bind:class="{ 'card-outline-danger': branchState === 'failure', 'card-outline-success': branchState === 'success' }">
            <div class="card-block">
                <h4 class="card-title">
                    <a v-bind:href="branch.commits_url" class="deco-none">{{ branch.name }}</a>
                </h4>

                <p class="card-text">
                    <span class="text-success" v-if="branchState === 'success'">
                        Success
                    </span>
                    <span class="text-danger" v-if="branchState === 'failure'">
                        Failure
                    </span>
                    <span v-if="branchState === 'pending'">
                    	Pending
                    </span>
                    <span v-if="branchState === 'no checks'">
                        No status
                    </span>
                    
                    <span>
                        &middot; <small class="text-muted"><span v-moment-ago="branch.last_updated"></span></small>
                    </span>
                    
                    &middot;
                    
                    <small class="code text-muted">
                        <a v-bind:href="branch.commit_url" class="deco-none">{{ truncatedSha }}</a>
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
