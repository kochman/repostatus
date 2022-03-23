app.component('index', {
	data() {
		return {
			orgRepo: "",
		}
	},
	methods: {
		getStatus() {
		    if (this.orgRepo === "") {
		        var org = "wtg";
		        var repo = "shuttletracker";
			} else {
                var split = this.orgRepo.split("/");
                if (split.length != 2) {
                    return
                }
                var org = split[0];
                var repo = split[1];
			}
			this.$router.push({ name: "repo", params: { org: org, repo: repo}});
		}
	},
	template: `
		<section class="jumbotron text-center">
			<div class="container">
				<h1 class="jumbotron-heading">RepoStatus</h1>
				<p class="lead text-muted">Get the status of any public GitHub repository's branches.</p>
				<form class="form-inline justify-content-center" v-on:submit.prevent="getStatus()">
					<div class="input-group">
						<input type="text" placeholder="wtg/shuttletracker" v-model="orgRepo" autocorrect="off" autocapitalize="off" spellcheck="false" class="form-control form-control-lg">
						<div class="input-group-btn">
							<button class="btn btn-lg btn-primary">Get status</button>
						</div>
					</div>
				</form>
			</div>
		</section>
	`
});
