document.addEventListener('keydown', function (event) {
    if (event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) {
        return;
    }

    if (event.key.toLowerCase() === 'p') {
        const footer = document.getElementById('footer');

        if (footer) {
            footer.toggleAttribute('hidden');
        }
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
