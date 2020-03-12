# Article workflow, CI/CD and troubleshooting

_Table of contents_

<!-- TOC depthFrom:2 insertAnchor:false orderedList:false updateOnSave:true withLinks:true -->

- [Article workflow, CI/CD and troubleshooting](#article-workflow-cicd-and-troubleshooting)
  - [Introduction](#introduction)
  - [Preparation of tools](#preparation-of-tools)
    - [Pre-commit](#pre-commit)
    - [black and blacken-docs](#black-and-blacken-docs)
    - [shfmt](#shfmt)
    - [Spell checking articles](#spell-checking-articles)
  - [Troubleshooting](#troubleshooting)
    - [Yaspeller](#yaspeller)
      - [New words on dictionary](#new-words-on-dictionary)
      - [Repeated words](#repeated-words)
    - [Prettier](#prettier)
    - [CI/CD](#cicd)
      - [Common](#common)
        - [Spell check](#spell-check)
        - [Link checker](#link-checker)
      - [Travis](#travis)
      - [Netlify](#netlify)
  - [Everything is fine](#everything-is-fine)

<!-- /TOC -->

## Introduction

Current article creation and development is defined in GUIDELINES in repository root and focus mostly on content creation, formatting in Markdown, image insertion, and styling.

Still, there are other pieces involved that are taken from `CI/CD` and could benefit from local 'checks' before waiting for CI to fail with not-so-clear messaging.

Note, that local preview of changes via the container or local Jekyll instance is also something to perform in addition to the checks described here.

## Preparation of tools

NOTE: We'll be executing the following commands from the repository checkout folder unless noted otherwise.

### Pre-commit

`pre-commit` is a set of hooks prepared to run on 'commit' time, ensuring that files are conformant to some good practices like:

- Executables have shebangs
- Proper end of file termination for text files (that includes `.svg` files too)
- Proper Markdown formatting with:
  - Extra lines before and after headings
  - Table columns sized to content
  - Remove duplicated spaces
  - etc.

We'll be installing `pre-commit` which setups a hook on git that is executed before, and then, use the defined configuration in the repository via `.pre-commit-config.yaml` to define the hook settings.

Check a list of available hooks at <https://pre-commit.com/hooks.html>, or course, others can be added outside of that list to the configuration so that they are applied to the new commits of the repository.

```sh
sudo dnf install pre-commit
pre-commit install
```

After it has been installed, a manual execution of pre-commit via `pre-commit run -a` will look somewhat similar to this:

```console
Check for useless excludes...............................................Passed
prettier.................................................................Passed
bashate..............................................(no files to check)Skipped
- hook id: bashate
black................................................(no files to check)Skipped
Check for added large files..............................................Passed
Check python ast.....................................(no files to check)Skipped
Check for case conflicts.................................................Passed
Check that executables have shebangs.................(no files to check)Skipped
Check JSON...............................................................Passed
Check for merge conflicts................................................Passed
Check for broken symlinks............................(no files to check)Skipped
Check vcs permalinks.....................................................Passed
Check Xml............................................(no files to check)Skipped
Check Yaml...............................................................Passed
Fix End of Files.........................................................Passed
Fix python encoding pragma...........................(no files to check)Skipped
Forbid new submodules....................................................Passed
Don't commit to branch...................................................Passed
Fix requirements.txt.................................(no files to check)Skipped
Sort simple YAML files...............................(no files to check)Skipped
Trim Trailing Whitespace.................................................Passed
Flake8...............................................(no files to check)Skipped
shfmt....................................................................Passed
blacken-docs.............................................................Passed
```

### black and blacken-docs

For Python files or Python snippets in documents, `black` will apply formatting to the code, ensuring indentation, proper usage of variables, spaces before and after operators, etc. Sometimes it might happen that an error is reported by the plugin, in that case, there's no general guideline, but it might be some error in the code like unterminated parenthesis, `"` or others.

### shfmt

Similar to `black`, `shfmt` is a tool written in Go that applies formatting recommendations to shell files. It will modify bash files to ensure indentation is set to 4 spaces, variables are used with `${}`, etc.

The tool must be available, so have `golang` packages installed and execute the [required steps](https://github.com/mvdan/sh) to install it, which, at this time is just running:

```sh
GO111MODULE=on go get mvdan.cc/sh/v3/cmd/shfmt
```

In case of any failure during the checks performed by `shfmt`, please review the `bash` file reported.

### Spell checking articles

Next step is to cover spell checking, so that we can locate usual typos:

- repeated words like `to to do xxx`
- word typos like `abstration` instead of `abstraction`
- etc

This is achieved via `yaspeller` tool and the `yaspeller-ci` hook.

We're going to install it using the `npm` installer so that it gets into the `node_modules` folder of the repository, and later, enable it to use with `pre-commit` (it installs a new hook overwriting the one from pre-commit so that both work together):

```sh
sudo dnf install npm
npm install --save-dev pre-commit
npm install --save-dev yaspeller-ci
pre-commit install -f
```

With above commands we're ready to spell check the repository contents by running `npm test` (as defined in `package.json`) or when using `git commit`.

## Troubleshooting

After the tools are in place, when a new commit is performed, `pre-commit` hook will perform several checks, the first one by `yaspeller-ci` will trigger a spell check on the files.

### Yaspeller

#### New words on dictionary

For each error, check correct word syntax, and if something is still ok, but reported, consider expanding the dictionary in `.yaspeller.json` to add a new word.

As a good practice, ensure that all words except last one contain a `,` afterwards, and also, use sorting so that next user can quickly find if a word is there or not, and add it.

You can use code editors with a 'trick' to maintain the file:

- just add a `,` to the last word
- allow editor to apply formatting so that words in same line are separated into individual, etc
- select the lines so that you perform sorting of the words
- finally, remove the `,` after the last word in dictionary

#### Repeated words

Yaspeller can also check for repeated words, this is a bit tricky, as it does so by removing code blocks, etc. Default settings we use, do ignore words with numbers and report repeated words, so it might happen that a sentence like:

    In case of the `Bare Metal Operator` the branch

Above sentence will report duplicated `the` because when `yaspeller` removes the code block, will just contain `the the` together.

It can be fixed by adding a `,` (which in fact, should be there)

    In case of the `Bare Metal Operator`, the branch

The hard part is to determine or find where it happened, probably easiest way is to start removing parts of the text on a temporary file, until the error is no reported and then, compare the last text that was removed to locate the error reported.

### Prettier

`Prettier` does check for formatting in Markdown, but in an 'opinionated' way... it will modify the Markdown so that it passes its preferences for formatting.

This is usually good, but sometimes with some lists that were not properly isolated from prior title or paragraph, it can be a bit messy.

As this is happening when the file has been added to a commit (but still not committed as we're running the commit command), you can use `git diff` to check the differences, validate and then `git add the-file` again.

### CI/CD

#### Common

There are some common tasks run by the CI/CD as part of the `Rakefile` like spell checking and `linkchecker`

##### Spell check

For references to spell check, please see [above section](#spell-checking-articles) about it when using `pre-commit`.

##### Link checker

Website articles and generated pages are validated for broken links via `linkchecker`.

`Linkchecker` does two tests:

1. Internal links checking
1. External links checking

To do so, a set of domains is defined as local and replaced to 'internal' links and the others are kept as they are. `Linkchecker` then, will start accessing each of those links and report their status if content is not retrieved.

Usual `linkchecker` failures involve `youtube` videos (that's why in the end we disabled their checking) and some internal links as well as git repositories.

Git repositories tend to get reorganized, moving documents around, etc. Best way to fix them, once they have failed, it to check the link in the repository, try to travel back a few commits, and use the resulting url that indicates the commit id so that it never changes again.

About internal links, there might be errors in manual defined links or sometimes because of errors in the website templates, for example, recently there was an issue when category/tag contain spaces, so that the link resulted in either "key word" or "key-word" depending on where it did appear, this resulted in unreachable pages that were reported as broken links in several pages at time.

Please note that `linkchecker` **checks** on the resulting generated website, because of it, the broken link, caused by one new blog post could end up being reported in `categories`, `blog post` or other `pages` that get generated during final website creation.

#### Travis

Travis configuration is defined in `.travis.yml` file at the repository root. It's read by the <https://travis-ci.org> servers and performs the actions listed there.

Note that `Travis` runs spell checking test, `Linkchecker` and page build.

There should be little errors (outside of transient package downloads uses for setting up the build environment).

If no new commit has been made to the website and there's a new reported issue in build, the `culprit` would be probably a broken external link.

If that's not the cause, check if some package defined in `package-lock.json` might be outdated as sometimes it might get 'broken' when depending packages are updated. We do keep it in the repository to have a more or less 'stable' environment for building the website versus having each time something completely different.

#### Netlify

Netlify is the service we use for rendering the PR's and it's doing the same job as Travis, so same troubleshooting tips and tricks apply.

Netlify does some intelligent caching of the generated files, so it's faster on build for the website, but it might cause some issues from time to time, so a 'clear-cache and rebuild' is required (which is done from their website interface for the failed build)

## Everything is fine

If local tests and CI tests have passed, you're ready... Travis will publish the website with the changes automatically when it gets merged into `source`. If for any reason, there are failures (typos, temporary errors, etc) the updated pages, even if merged, will not get published.

![](https://media1.tenor.com/images/4d1f1b713a22c2de9da8f428bff50a29/tenor.gif)

Remember that last step of all validation is to push to Github pages, without that step, no rendering is done.

This affects specially things that tend to get broke like URL's, etc if a URL is broken, a commit is needed to fix them before other changes can pass CI and get `ACK` for merging, and of course, a `rebase` of the pending ones on top of the fixed version.
