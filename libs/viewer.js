const source = new EventSource("/events");

source.addEventListener("open", () => {
    console.log("SSE connected");
});

source.addEventListener("slide-changed", (event) => {
    console.log("Received slide:", event.data);

    const params = new URLSearchParams({
        slide: event.data,
        presenter: "false"
    });

    window.location.href = `/slides?${params.toString()}`;
});

source.addEventListener("error", () => {
    console.log("SSE connection error; readyState:", source.readyState);
});