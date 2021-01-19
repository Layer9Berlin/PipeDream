![PipeDream](docs/assets/PipeDream.png)

>## 🚧 Status: alpha 🚧
>#### Early development, feedback wanted

# PipeDream

PipeDream is an open-source, general-purpose **automation tool**.

It is an alternative to shell scripts - just as powerful, but more **composable**, **testable** and **reliable**.

## How it works

Workflows - called pipelines - are defined in yaml files. Using a simple, but extensible syntax, you can call, chain, merge, and split pipelines, invoke shell commands, handle errors, use a key-value store for intermediate results, write tests, create mocks, and much more.

## What it does

Shell scripting is an art that takes years to master. Tools like [shellcheck](https://github.com/koalaman/shellcheck) and [bats](https://github.com/bats-core/bats-core) can help you improve the code quality of your scripts, but they don't solve the core problems: 
- There are many traps for beginners to fall into
- Each command has its own syntax (not to mention subtle differences between shells)
- Error handling is so difficult it's rarely done well
- Even simple tasks often require contorted solutions that break for edge cases
- A lack of tests and documentation means you generally don't want to adapt an existing script

If you have used a modern CI/CD, orchestration or containerization tool, you are probably already familiar with a different way to write a script: defining tasks in yaml format. PipeDream is the natural extension of these tools to your localhost. Automate anything, anywhere, using a simple syntax that follows the principle of least astonishment.

> _"It's like Ansible for localhost!"_
>
> Anonymous

We have created PipeDream to simplify maintenance and dependency management tasks that used to involve a large number of steps, combining many different tools running both locally and on remote servers.

To illustrate its potential, we have used PipeDream to define a **universal dependency manager** pipe. It offers a single, consistent interface pulling together results from different package managers.

## Try it out

### Installation

From `npm`

```npm i @l9/pipedream```

Using `homebrew` (Mac OS)

```
brew tap https://github.com/Layer9Berlin/pipedream
brew install pipedream
```

Using `apt-get`
```
echo "deb https://github.com/Layer9Berlin/pipedream $(lsb_release -cs) main" | sudo tee -a /etc/apt/sources.list
sudo apt-get update && sudo apt-get upgrade
sudo apt-get install pipedream
```

From source (requires Go installed)

```
git clone https://github.com/Layer9Berlin/pipedream
cd pipedream
bin/bootstrap
```
then select `Installation` and `Install` when promted.

This will run a compiled version of PipeDream, allowing you to execute the `installation.pipe` file (very meta), which in turn compiles your local directory and installs everything you need. If you make changes to the code, just run PipeDream and execute the `installation.pipe` again to update your installed version.

### Write your first pipe

Check out the [Quick Start Guide](./docs/quick-start) or the [Documentation](./cmd). You can also have a look at the





Software developers and system admins have no shortage of tools to automate their workflows. When processes require a number of different tools, however, they often slip through the cracks and are either performed manually, with flaky shell scripts or not at all.

For example, simply keeping a modern website up-to-date involves many different tools:
-   ##### Programming language runtime
    Possibly with its own package manager like rvm or nvm.
-   ##### Web framework
-   ##### One package manager per programming language
    Itself a versioned piece of software
-   ##### Several server environments 
    (each with their own versions of installed dependencies)
-   ##### Docker containers
    (again, each with their own versions)
-   ##### other locally installed software
    partly managed by a package manager
-   ##### OS-level server software
    with its own package manager
-   ##### Version Control System
    like git (through which new versions might also become available)
-   ##### a compiler, a linter and a testing framework each with a peer dependency on the programming language
-   ##### CI/CD pipeline
    with many implicit dependencies in configuration files and shell scripts
-   ##### Server orchestration tools
-   ##### Deployment tools
and probably many more.

The cross-dependencies between all these tools are usually implicit and problems are fixed by hand as they arise.

Making these relationships explicit and establishing simple processes for 


While some steps - like checking changelogs and adapting to breaking changes - need to be performed by hand, 


This also means that processes like upgrading dependencies are far from the



Unfortunately, this also means that 

PipeDream aims to fill the gap between shell scripts, manual processes and requirements

Automation of workflows 

Workflow automation 

- ##### Comprehensive overview of current repo state
    - ###### Git status
        Are there untracked changes, commit to push, or upstream changes to pull?
    - ###### Updates
        Are new versions of npm packages available? What about Node or npm itself?
    - ###### Vulnerabilities
    - ###### Version inconsistencies:
        Does the locally installed version match the one deployed on each server?
        Does it match the one configured for the Docker container?
        Are there other mismatched peer dependencies e.g. in linter or deployment config? 

- ##### Outdated dependencies compiled into a software deployed to a server
    - Check deployed software version
    - 
    - Spin up Docker container with outdated dependencies
    - 


It is common for 

It can be used to combine scripts defined with different 


- reliable
- composable
- testable
- shareable
- maintainable

## What is it for?



error handling
testability

## Getting started


## TODO

See [issues](/../../issues)
 
