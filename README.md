# vll - Visual Linear Logic

This is an implementation of the visual linear logic notation I described here http://adelelopez.com/visual-linear-logic, which is based on this paper by Brady and Trimble: https://core.ac.uk/download/pdf/82545173.pdf

Right now it's just a proof-of-concept, with only the bare minimum functionality to be a proof editor. Only the multiplicative subset of linear logic is implemented.

I was inspired to try making this an actual editor after I saw this https://github.com/peterhellberg/pixel-experiments/tree/master/metaballs and saw that it was easier than expected to implement the "blobby" behavior I wanted it to have.

## Controls
I've tried to make the controls relatively intuitive. You start out in create mode, which lets you right click to add a new bubble (of the opposite color), or press a character to create a new bubble with that variable name (space creates a new unit of the same color).
You can press backspace or delete to delete any bubbles, and you can drag-and-drop bubbles into each other. The titlebar shows your statement in traditional (Tolestra's) notation.
Once you've finished creating your initial statement, you can press enter to go into proof mode.

Once in proof mode, you can't (barring any bugs) do any manipulations which are logically incorrect. Space still lets you create new units, and tab lets you nest your bubble in a loop of the opposite color.
Drag-and-drop now only works when it is logically correct, and right-click drag-and-drop creates a new assumption pair, which are shown as a yellow and purple bubble. These bubbles can have new variables added to them, but putting a variable in one will put it in the other.
If you want to assume more complicated statements, it is possible but you have to be clever about what order you do things in.

At any time, you can grab a bubble to move it, and you can jerk a grabbed bubble to detach it from its parent, so you can move it somewhere else (but it will snap back to its original place if that's not allowed).

## Roadmap
Right now the code isn't especially great, and needs much more testing before I'd really be comfortable counting on its logical rigor.

I'm working on this in my spare time, and according to how excited/motivated I feel about it. If you want me to implement something sooner, just let me know by creating an issue, and that will encourage me to do it!

### Quality of life
I plan to add more comprehensive testing, and to refactor things to be cleaner/faster.

I'm planning to add more features to the sidebar, so that you can see the whole list of steps you have taken to get to this point, and can move between them if desired, just like a real proof assistant.

Also, I'll add a quad-tree (and batching) to make drawing more efficient, and figure out a way to have bubbles move out of each other's way better.

### New logic features

The next feature I plan to add will be the exponential operators.

After that, the additive connectives (additive units might be a bit later, since I haven't designed notation for them yet).

Then, I'll implement quantifiers so that it can do first-order logic (the notation for this is designed, even though it's not in the blog post).

Beyond that, I'm not sure. I might go for second-order logic / type theory, but that would require more design work first.

