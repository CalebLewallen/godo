![godo-gopher](/assets/static/godo-gopher.png)

# Project
GoDo is a todo application written in Go. I wasn't happy with the todo apps I had access to at work, so I made this one. I'll keep refining it, but it's setup for how I like to use TUI apps, which may be weird to you.

## Interface
The interface should be fairly simple. A sidebar on the left with folders and todo's. Folders are optional. The label for each should by the todo name or the folder name.

Use icons and font color to differentiate them. The sidebar should have tabs to show completed and incomplete tasks.

```
My Todo's
---
/ Folder
  * todo
  * todo
```
```
Completed Tasks
---
✓ todo
✓ todo
```

The main content area of the page should have places to enter the following information.
- Task Name
- Due Date
- Folder
- Description

The description field is Markdown. I'd like to have syntax highlighting in the description.

## Interactions

This should be primarily keyboard driven, but should have mouse support enabled.

### Global
ctrl + e: focus between sidebar and task
ctrl + b: open/close the sidebar
ctrl + f: find folders and tasks (filter the sidebar items)
ctrl + q: quit the application
ctrl + n: new task (in active folder, based on sidebar or currently open task)
alt + ctrl + n: new folder (in active folder, or based on currently open task)

### Sidebar
shift + tab: switch tabs
up/down arrows: highlight items visible in the tree
right arrow: if on a folder, expand the folder
left arrow: if on a folder, collapse the folder
enter: if on a folder, expand/collapse the folder. if on a task, open the task and focus on the task

### Task
tab: go to next field
shift + tab: go to previous field
ctrl + s: save the task
ctrl + w: close task (ask to (s) save task, (d)discard changes, or (c) cancel if there are unsaved changes)

### Data
Save the data to a sqlite database

## Install
As a Go application, you should be able to `go install` and open the app with `godo`.

## Updates
I want to be able to pull the latest version from github with `godo --update`. This should update the application and run and database migrations.

## Uninstall
Users should be able to remove the program and associated sqlite file with `godo --uninstall`



