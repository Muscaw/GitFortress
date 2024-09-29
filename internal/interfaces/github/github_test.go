package github

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type requestInformation struct {
	requestUri string
	headers    http.Header
}

func Test_list_owned_repositories_no_repositories_available(t *testing.T) {
	var ri *requestInformation = nil
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ri = &requestInformation{requestUri: r.RequestURI, headers: r.Header}
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	github, err := GetGithubVCS(testServer.URL, "some-token")
	if err != nil {
		t.FailNow()
	}

	repos, err := github.ListOwnedRepositories()
	if err != nil {
		t.FailNow()
	}

	if ri == nil {
		t.FailNow()
	}

	if len(repos) != 0 {
		t.Fatalf("expected 0 repos, got %v", len(repos))
	}

	if ri.headers.Get("Authorization") != "Bearer some-token" {
		t.Fatalf("expected authorization header Bearer some-token. got %v", ri.headers.Get("Authorization"))
	}
}

const someRepos string = `
[
  {
    "id": 1296269,
    "node_id": "MDEwOlJlcG9zaXRvcnkxMjk2MjY5",
    "name": "Hello-World",
    "full_name": "octocat/Hello-World",
    "owner": {
      "login": "octocat",
      "id": 1,
      "node_id": "MDQ6VXNlcjE=",
      "avatar_url": "https://github.com/images/error/octocat_happy.gif",
      "gravatar_id": "",
      "url": "https://api.github.com/users/octocat",
      "html_url": "https://github.com/octocat",
      "followers_url": "https://api.github.com/users/octocat/followers",
      "following_url": "https://api.github.com/users/octocat/following{/other_user}",
      "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
      "organizations_url": "https://api.github.com/users/octocat/orgs",
      "repos_url": "https://api.github.com/users/octocat/repos",
      "events_url": "https://api.github.com/users/octocat/events{/privacy}",
      "received_events_url": "https://api.github.com/users/octocat/received_events",
      "type": "User",
      "site_admin": false
    },
    "private": false,
    "html_url": "https://github.com/octocat/Hello-World",
    "description": "This your first repo!",
    "fork": false,
    "url": "https://api.github.com/repos/octocat/Hello-World",
    "archive_url": "https://api.github.com/repos/octocat/Hello-World/{archive_format}{/ref}",
    "assignees_url": "https://api.github.com/repos/octocat/Hello-World/assignees{/user}",
    "blobs_url": "https://api.github.com/repos/octocat/Hello-World/git/blobs{/sha}",
    "branches_url": "https://api.github.com/repos/octocat/Hello-World/branches{/branch}",
    "collaborators_url": "https://api.github.com/repos/octocat/Hello-World/collaborators{/collaborator}",
    "comments_url": "https://api.github.com/repos/octocat/Hello-World/comments{/number}",
    "commits_url": "https://api.github.com/repos/octocat/Hello-World/commits{/sha}",
    "compare_url": "https://api.github.com/repos/octocat/Hello-World/compare/{base}...{head}",
    "contents_url": "https://api.github.com/repos/octocat/Hello-World/contents/{+path}",
    "contributors_url": "https://api.github.com/repos/octocat/Hello-World/contributors",
    "deployments_url": "https://api.github.com/repos/octocat/Hello-World/deployments",
    "downloads_url": "https://api.github.com/repos/octocat/Hello-World/downloads",
    "events_url": "https://api.github.com/repos/octocat/Hello-World/events",
    "forks_url": "https://api.github.com/repos/octocat/Hello-World/forks",
    "git_commits_url": "https://api.github.com/repos/octocat/Hello-World/git/commits{/sha}",
    "git_refs_url": "https://api.github.com/repos/octocat/Hello-World/git/refs{/sha}",
    "git_tags_url": "https://api.github.com/repos/octocat/Hello-World/git/tags{/sha}",
    "git_url": "git:github.com/octocat/Hello-World.git",
    "issue_comment_url": "https://api.github.com/repos/octocat/Hello-World/issues/comments{/number}",
    "issue_events_url": "https://api.github.com/repos/octocat/Hello-World/issues/events{/number}",
    "issues_url": "https://api.github.com/repos/octocat/Hello-World/issues{/number}",
    "keys_url": "https://api.github.com/repos/octocat/Hello-World/keys{/key_id}",
    "labels_url": "https://api.github.com/repos/octocat/Hello-World/labels{/name}",
    "languages_url": "https://api.github.com/repos/octocat/Hello-World/languages",
    "merges_url": "https://api.github.com/repos/octocat/Hello-World/merges",
    "milestones_url": "https://api.github.com/repos/octocat/Hello-World/milestones{/number}",
    "notifications_url": "https://api.github.com/repos/octocat/Hello-World/notifications{?since,all,participating}",
    "pulls_url": "https://api.github.com/repos/octocat/Hello-World/pulls{/number}",
    "releases_url": "https://api.github.com/repos/octocat/Hello-World/releases{/id}",
    "ssh_url": "git@github.com:octocat/Hello-World.git",
    "stargazers_url": "https://api.github.com/repos/octocat/Hello-World/stargazers",
    "statuses_url": "https://api.github.com/repos/octocat/Hello-World/statuses/{sha}",
    "subscribers_url": "https://api.github.com/repos/octocat/Hello-World/subscribers",
    "subscription_url": "https://api.github.com/repos/octocat/Hello-World/subscription",
    "tags_url": "https://api.github.com/repos/octocat/Hello-World/tags",
    "teams_url": "https://api.github.com/repos/octocat/Hello-World/teams",
    "trees_url": "https://api.github.com/repos/octocat/Hello-World/git/trees{/sha}",
    "clone_url": "https://github.com/octocat/Hello-World.git",
    "mirror_url": "git:git.example.com/octocat/Hello-World",
    "hooks_url": "https://api.github.com/repos/octocat/Hello-World/hooks",
    "svn_url": "https://svn.github.com/octocat/Hello-World",
    "homepage": "https://github.com",
    "language": null,
    "forks_count": 9,
    "stargazers_count": 80,
    "watchers_count": 80,
    "size": 108,
    "default_branch": "master",
    "open_issues_count": 0,
    "is_template": true,
    "topics": [
      "octocat",
      "atom",
      "electron",
      "api"
    ],
    "has_issues": true,
    "has_projects": true,
    "has_wiki": true,
    "has_pages": false,
    "has_downloads": true,
    "archived": false,
    "disabled": false,
    "visibility": "public",
    "pushed_at": "2011-01-26T19:06:43Z",
    "created_at": "2011-01-26T19:01:12Z",
    "updated_at": "2011-01-26T19:14:43Z",
    "permissions": {
      "admin": false,
      "push": false,
      "pull": true
    },
    "allow_rebase_merge": true,
    "template_repository": null,
    "temp_clone_token": "ABTLWHOULUVAXGTRYU7OC2876QJ2O",
    "allow_squash_merge": true,
    "allow_auto_merge": false,
    "delete_branch_on_merge": true,
    "allow_merge_commit": true,
    "subscribers_count": 42,
    "network_count": 0,
    "license": {
      "key": "mit",
      "name": "MIT License",
      "url": "https://api.github.com/licenses/mit",
      "spdx_id": "MIT",
      "node_id": "MDc6TGljZW5zZW1pdA==",
      "html_url": "https://github.com/licenses/mit"
    },
    "forks": 1,
    "open_issues": 1,
    "watchers": 1
  },
  {
    "id": 1296269,
    "node_id": "MDEwOlJlcG9zaXRvcnkxMjk2MjY5",
    "name": "other-repo",
    "full_name": "octocat/other-repo",
    "owner": {
      "login": "octocat",
      "id": 1,
      "node_id": "MDQ6VXNlcjE=",
      "avatar_url": "https://github.com/images/error/octocat_happy.gif",
      "gravatar_id": "",
      "url": "https://api.github.com/users/octocat",
      "html_url": "https://github.com/octocat",
      "followers_url": "https://api.github.com/users/octocat/followers",
      "following_url": "https://api.github.com/users/octocat/following{/other_user}",
      "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
      "organizations_url": "https://api.github.com/users/octocat/orgs",
      "repos_url": "https://api.github.com/users/octocat/repos",
      "events_url": "https://api.github.com/users/octocat/events{/privacy}",
      "received_events_url": "https://api.github.com/users/octocat/received_events",
      "type": "User",
      "site_admin": false
    },
    "private": false,
    "html_url": "https://github.com/octocat/other-repo",
    "description": "This your first repo!",
    "fork": false,
    "url": "https://api.github.com/repos/octocat/other-repo",
    "archive_url": "https://api.github.com/repos/octocat/other-repo/{archive_format}{/ref}",
    "assignees_url": "https://api.github.com/repos/octocat/other-repo/assignees{/user}",
    "blobs_url": "https://api.github.com/repos/octocat/other-repo/git/blobs{/sha}",
    "branches_url": "https://api.github.com/repos/octocat/other-repo/branches{/branch}",
    "collaborators_url": "https://api.github.com/repos/octocat/other-repo/collaborators{/collaborator}",
    "comments_url": "https://api.github.com/repos/octocat/other-repo/comments{/number}",
    "commits_url": "https://api.github.com/repos/octocat/other-repo/commits{/sha}",
    "compare_url": "https://api.github.com/repos/octocat/other-repo/compare/{base}...{head}",
    "contents_url": "https://api.github.com/repos/octocat/other-repo/contents/{+path}",
    "contributors_url": "https://api.github.com/repos/octocat/other-repo/contributors",
    "deployments_url": "https://api.github.com/repos/octocat/other-repo/deployments",
    "downloads_url": "https://api.github.com/repos/octocat/other-repo/downloads",
    "events_url": "https://api.github.com/repos/octocat/other-repo/events",
    "forks_url": "https://api.github.com/repos/octocat/other-repo/forks",
    "git_commits_url": "https://api.github.com/repos/octocat/other-repo/git/commits{/sha}",
    "git_refs_url": "https://api.github.com/repos/octocat/other-repo/git/refs{/sha}",
    "git_tags_url": "https://api.github.com/repos/octocat/other-repo/git/tags{/sha}",
    "git_url": "git:github.com/octocat/other-repo.git",
    "issue_comment_url": "https://api.github.com/repos/octocat/other-repo/issues/comments{/number}",
    "issue_events_url": "https://api.github.com/repos/octocat/other-repo/issues/events{/number}",
    "issues_url": "https://api.github.com/repos/octocat/other-repo/issues{/number}",
    "keys_url": "https://api.github.com/repos/octocat/other-repo/keys{/key_id}",
    "labels_url": "https://api.github.com/repos/octocat/other-repo/labels{/name}",
    "languages_url": "https://api.github.com/repos/octocat/other-repo/languages",
    "merges_url": "https://api.github.com/repos/octocat/other-repo/merges",
    "milestones_url": "https://api.github.com/repos/octocat/other-repo/milestones{/number}",
    "notifications_url": "https://api.github.com/repos/octocat/other-repo/notifications{?since,all,participating}",
    "pulls_url": "https://api.github.com/repos/octocat/other-repo/pulls{/number}",
    "releases_url": "https://api.github.com/repos/octocat/other-repo/releases{/id}",
    "ssh_url": "git@github.com:octocat/other-repo.git",
    "stargazers_url": "https://api.github.com/repos/octocat/other-repo/stargazers",
    "statuses_url": "https://api.github.com/repos/octocat/other-repo/statuses/{sha}",
    "subscribers_url": "https://api.github.com/repos/octocat/other-repo/subscribers",
    "subscription_url": "https://api.github.com/repos/octocat/other-repo/subscription",
    "tags_url": "https://api.github.com/repos/octocat/other-repo/tags",
    "teams_url": "https://api.github.com/repos/octocat/other-repo/teams",
    "trees_url": "https://api.github.com/repos/octocat/other-repo/git/trees{/sha}",
    "clone_url": "https://github.com/octocat/other-repo.git",
    "mirror_url": "git:git.example.com/octocat/other-repo",
    "hooks_url": "https://api.github.com/repos/octocat/other-repo/hooks",
    "svn_url": "https://svn.github.com/octocat/other-repo",
    "homepage": "https://github.com",
    "language": null,
    "forks_count": 9,
    "stargazers_count": 80,
    "watchers_count": 80,
    "size": 108,
    "default_branch": "master",
    "open_issues_count": 0,
    "is_template": true,
    "topics": [
      "octocat",
      "atom",
      "electron",
      "api"
    ],
    "has_issues": true,
    "has_projects": true,
    "has_wiki": true,
    "has_pages": false,
    "has_downloads": true,
    "archived": false,
    "disabled": false,
    "visibility": "public",
    "pushed_at": "2011-01-26T19:06:43Z",
    "created_at": "2011-01-26T19:01:12Z",
    "updated_at": "2011-01-26T19:14:43Z",
    "permissions": {
      "admin": false,
      "push": false,
      "pull": true
    },
    "allow_rebase_merge": true,
    "template_repository": null,
    "temp_clone_token": "ABTLWHOULUVAXGTRYU7OC2876QJ2O",
    "allow_squash_merge": true,
    "allow_auto_merge": false,
    "delete_branch_on_merge": true,
    "allow_merge_commit": true,
    "subscribers_count": 42,
    "network_count": 0,
    "license": {
      "key": "mit",
      "name": "MIT License",
      "url": "https://api.github.com/licenses/mit",
      "spdx_id": "MIT",
      "node_id": "MDc6TGljZW5zZW1pdA==",
      "html_url": "https://github.com/licenses/mit"
    },
    "forks": 1,
    "open_issues": 1,
    "watchers": 1
  }
]
`

func Test_list_owned_repositories_some_repositories_available(t *testing.T) {
	var ri *requestInformation
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ri = &requestInformation{requestUri: r.RequestURI, headers: r.Header}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(someRepos))
	}))
	defer testServer.Close()

	github, err := GetGithubVCS(testServer.URL, "some-token")
	if err != nil {
		t.FailNow()
	}

	repos, err := github.ListOwnedRepositories()
	if err != nil {
		t.FailNow()
	}

	if ri == nil {
		t.FailNow()
	}

	if len(repos) != 2 {
		t.Fatalf("expected 0 repos, got %v", len(repos))
	}

	if ri.headers.Get("Authorization") != "Bearer some-token" {
		t.Fatalf("expected authorization header Bearer some-token. got %v", ri.headers.Get("Authorization"))
	}
}
