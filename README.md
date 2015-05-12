# git env

A tool based on ENV branching as described here: http://www.wearefine.com/mingle/env-branching-with-git/

## Installation

If you have Go installed and set up:

    go get github.com/dxw/git-env

If you prefer to download a binary: https://github.com/dxw/git-env/releases

## Usage

Start by answering a couple of questions about how your repository is set up:

    git env init

### `git env start feature/meow`

If using the defaults options, this is equivalent to:

    git checkout master
    git pull --rebase origin master
    git checkout -b feature/meow

### `git env deploy stage feature/meow`

(You may omit the second argument to deploy if you want to merge the current branch).

This is equivalent to:

    git checkout feature/meow
    git pull --rebase origin master
    git checkout stage
    git pull --rebase origin stage
    git merge feature/meow

### `git env deploy master feature/meow`

This is equivalent to:

    git checkout feature/meow
    git pull --rebase origin master
    git checkout master
    git merge --no-ff feature/meow

## Notes

* The original document mentions pushing your changes after merging. The git-env tool will never push.
* git-env runs commands like "git pull --rebase origin master" - it will use the value of `branch.master.remote` for the remote in that command.
