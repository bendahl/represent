# Represent - A simple Markdown Presentation Tool
This is a simple standalone tool to quickly create a presentation based on Markdown files.
The workflow is simple:

* Write your slides as a single Markdown file.
* Generate slides to a separate folder.
* Start the presentation mode.
* Open up a browser window and point it at [the slides endpoint](http://localhost:8080/slides).
* _Optional:_ Open up a second browser window (preferably on a different screen) to show the presenter notes 
and point it at [the presenter endpoint](http://localhost:8080/present).
* To move back and forth through the slides, simply use the right and left arrow keys on your keyboard.
  * **Note: Moving between slides is synced between endpoints, so pressing the arrow keys on either view (slides or presenter) works.**

## Usage
As mentioned above, there are different phases when creating a presentation:
1. Markdown creation 
2. Slide generation
3. Presenting

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
  * Use `:start-notes` and `:end-notes` markers to mark a block of presenter notes
  * Presenter notes can make use of the same formatting as any regular slide

For more details see the `example/example.md` file in this repository. It contains examples for the above-mentioned features and is used for basic testing.

### Slide Generation
To generate the slides, simply use the follwing command:

> represent -theme [light|dark] -out [slide-folder] markdown.md

**Note that -theme and -out are optional parameters. Dark is used as the default theme. If no output directory name 
is chosen, `slides/` will be assumed as the target directory. The directory will be created if it does not exist.**

### Presentation
To present, simply run the below command:

> represent -present ./slide-folder

The slides will be published on the endpoint http://localhost:8080/slides.
The presenter view will be published on the endpoint http://localhost:8080/present.
Use right and left arrow keys to navigate between slides.
You can use the arrow keys on the presenter view as well as on the slides.

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