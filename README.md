# Represent - A simple Markdown Presentation Tool
This is a simple standalone tool to quickly create a presentation based on Markdown files.
The workflow is simple:

* Write your slides as a single Markdown file.
* Start the presentation server, pointing it at your Markdown.
* Open up a browser window and point it at [the presenter endpoint](http://localhost:8080/presenter).
* Open up a second browser window at [the slides endpoint](http://localhost:8080/slides)
* Done!

**IMPORTANT:** 
- You can navigate back and forth between slides using the left and right arrow keys.
- Navigation only works in the presenter view.
- The slides view is kept in sync automatically via SSE (server sent events).
- Notes are only shown in the presenter view and can be toggled via the p-key (p stands for "presenter notes"). 

### Markdown Creation
The most important thing to keep in mind when creating the slides is that they all go in one single .md-file.
Besides this, here are some things to note:

* The title of your presentation is marked by a h1 header (`#` in Markdown)
* Any lines that come after the title are interpreted as the subtitle and will be converted to h3 (`###`)
* Subsequent (regular slides) should start with a h2 header (`##`)
* Tables are supported
* Code blocks are supported (including syntax highlighting)
* Light and dark mode are supported (see [slide generation](#slide-generation) for details)
* Assets, like images, should be copied to a separate folder in the working directory. The folder needs to be called `assets`.
* Presenter notes are supported
  * Use `:notes-start` and `:notes-end` markers to mark a block of presenter notes
  * Presenter notes can make use of the same formatting as any regular slide

For more details see the `example/example.md` file in this repository. It contains examples for the above-mentioned features and is used for basic testing.

### Presentation
To present, simply run the below command:

> represent -present -in ./example/example.md

The slides will be published on the endpoint http://localhost:8080/slides.
Use right and left arrow keys to navigate between slides.
Toggle notes via the "p" key (p stands for presenter, like in "presenter mode").

### Slide Rendering
To render slides to HTML files, simply use the following command:

> ./represent render -in ./example/example.md -out example/slides

**Note that this is an optional step. You can render to slides to HTML if you like, but for a presentation this is not required.
At this point, the same templates are used for on-the-fly presenting as well as local rendering. 
This may change in the future to optimize the output for different purposes.**

## Development Status and AI Assistance
This project was developed iteratively with assistance from ChatGPT and other AI-assisted coding techniques. 
The implementation reflects an evolving design, and the code may contain inconsistencies, technical debt, 
or areas that would benefit from further refactoring.

The project is provided as-is, primarily for learning, experimentation, and community use. Testing, 
documentation, and code quality improvements may be undertaken over time, 
but no particular maintenance schedule or future improvements are guaranteed.

## Disclaimer
This software is provided "AS IS", without warranties or guarantees of any kind, to the extent permitted 
by applicable law. Use it at your own risk. The author does not guarantee that the software is error-free, 
secure, fit for a particular purpose, or suitable for production use.

Users are responsible for reviewing, testing, and validating the software before using it, particularly in environments involving sensitive data, credentials, or critical systems.

Nothing in this disclaimer excludes or limits liability to the extent such exclusion or limitation is prohibited by applicable law.

## Contributing
At this point in time I'm not accepting any pull requests. Neither will I accept any feature nor change requests.
Feel free to fork and modify this program as you wish, hoẃever.