const NOTES_HIDDEN_KEY = 'presenterNotesHidden';

// Restore the saved state, defaulting to visible notes.
let notesHidden = sessionStorage.getItem(NOTES_HIDDEN_KEY) === 'true';

function applyNotesVisibility() {
    const footer = document.getElementById('footer');

    if (footer) {
        footer.toggleAttribute('hidden', notesHidden);
    }
}

// Apply the saved state when the page is first loaded.
document.addEventListener('DOMContentLoaded', applyNotesVisibility);

// Reapply the state after htmx updates the slide or footer.
document.addEventListener('htmx:afterSettle', applyNotesVisibility);
document.addEventListener('htmx:oobAfterSwap', applyNotesVisibility);

document.addEventListener('keydown', function (event) {
    if (event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) {
        return;
    }

    if (event.key.toLowerCase() === 'p') {
        notesHidden = !notesHidden;

        sessionStorage.setItem(
            NOTES_HIDDEN_KEY,
            String(notesHidden)
        );

        applyNotesVisibility();
        return;
    }

    if (event.key === 'ArrowRight') {
        event.preventDefault();
        htmx.trigger('#next-slide', 'nextSlide');
    } else if (event.key === 'ArrowLeft') {
        event.preventDefault();
        htmx.trigger('#prev-slide', 'prevSlide');
    }
});
