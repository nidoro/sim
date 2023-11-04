# sim &mdash; Simple Discrete Event Simulation Framework in Go

> **Warning**
> This is a work-in-progress. There are currently no stable versions, and the latest code may be broken. Use it at your own risk.

## Introduction

The **sim** package provides a framework through which you can
implement discrete event simulation models.

My motivation to start this project is two-fold. First, I have
used a simulation framework in another language that I didn't like
because it didn't help me as much as I wished it did. For instance,
I had to calculate queue statistics myself, which was not trivial to do.
It is possible that I was not using the framework right, but in any case,
I have other reasons to start this project.

Secondly, although I'm not an Rockwell Arena hard-core user, I did enjoy
using it for small and medium-sized projects. However, I noticed that
creating large models via a graphical interface requires too much effort.
You can't do a for-loop in an interface to create multiple model blocks.

Thirdly, I wanted to practice my Go skills, a language that I've recently
started to enjoy, as well as my simulation skills. So I don't recommend
using this project for commercial purposes (at least not yet),
because I'm still learning.

## Overview

**sim** uses concepts present in other simulation software, like Arena:

- `sim.Entity`: a persistent, "living" object of the simulation. It moves around
through queues and processes, until it leaves the system.
- `sim.EntitySource`: an entity generator.
- `sim.Resource`: something that is required by processes.
- `sim.Process`: a process though which an entity can go though.
    
## Processes

Conceptually, a process can be one of the following:

- **Seize, delay, release**: The process requires one or more resources to
be executed, which are seized at the start of the process.
It takes some time to be completed and, after completion,
the resource(s) are released.
- **Seize, delay**: Similar to the previous one, but the resources
are not released after the execution. Instead, they are consumed.
- **Delay**: The process requires no resources. It just takes some time
to be completed.

### To be continued...


