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

### Features
- Create on-the-fly presentations from your Markdown files
- Separate presenter mode and viewer mode (synced via SSE)
- Switch between light and dark themes
- Use custom themes (via custom css)
- Render slides to HTML files

### Markdown Creation - Preparing the Presentation
The most important thing to keep in mind when creating the slides is that they all go in one single .md-file.
Besides this, here are some things to note:

* The title of your presentation is marked by a h1 header (`#` in Markdown)
* Any lines that come after the title are interpreted as the subtitle and will be converted to h3 (`###`)
* Subsequent (regular slides) should start with a h2 header (`##`)
* Tables are supported
* Code blocks are supported (including syntax highlighting)
* Light and dark mode are supported (see [slide generation](#slide-generation) for details)
* Assets, like images, should be copied to a separate folder in the working directory. The folder needs to be called `assets`.
  * **If you would like to customize the stylesheets for syntax highlighting or the slides, your custom css goes into the asset folder as well.
      Note that you only need to pass the filename of your css files as a program parameter, however.**
* Presenter notes are supported
  * Use `:notes-start` and `:notes-end` markers to mark a block of presenter notes
  * Presenter notes can make use of the same formatting as any regular slide

For more details see the `example/example.md` file in this repository. It contains examples for the above-mentioned features and is used for basic testing.

### Presenting
The presentation can be started by the following command:

> ./represent present -in <path/to/md/file> -host <hostname|IP> -port <portnumber>

**Note that hostname and portname are optional parameters.**

#### Standard Endpoints
The default endpoints for side and presenter view are as follows:

- Presenter View: http://127.0.0.1:8080/presenter
- Slide View: http://127.0.0.1:8080/slides

### Theming
The default theme is a light theme along with a Github-light inspired HighlightJS theme for syntax highlighting. 
A default dark mode is included. the following two calls show how to switch between them explicitly:

**Use Default Dark Mode:**
> ./represent present -in example/example.md -theme dark -hl-theme dark

**Use Default Light Mode:**
> ./represent present -in example/example.md -theme light -hl-theme light

**Mix Default Dark and Light Mode:**
> ./represent present -in example/example.md -theme dark -hl-theme light

In addition, you may pass your own custom css. The structure should match the structure of the example css-files found in `example/assets`.
The file `purebasic-hljs.css` is a regular HighlightJS style. 
Any HighlightJS style can be applied for syntax highlighting.
The `purebasic.css` file shows the structure of the css-file applied to the slides.
Make sure to preserve this structure when creating custom themes.
Below is an example call of the program passing custom css-files.
**Note that your css-files go into the assets-folder, as shown in the example.**

**Using custom css:**
> ./represent present -in example/example.md -theme purebasic.css -hl-theme purebasic-hljs.css


### Slide Rendering
To render slides to HTML files, simply use the following command:

> ./represent render -in ./example/example.md -out example/slides

**Note that this is an optional step. You can render to slides to HTML if you like, but for a presentation this is not required.
At this point, the same templates are used for on-the-fly presenting as well as local rendering. 
This may change in the future to optimize the output for different purposes.**

## Development Status and AI Assistance
This project started as a little experiment to see how quickly a simple Markdown presentation tool could be created using AI assistance.
The result was a mostly functional but hard to maintain first version.
A full rewrite followed.
The current version has all basic features that were initially envisioned and is much easier to maintain.
Further improvements to the code structure and more tests will be added as time permits.

## Contributing
At this point in time I'm not accepting any pull requests. Neither will I accept any feature nor change requests.
Feel free to fork and modify this program as you wish, hoẃever.