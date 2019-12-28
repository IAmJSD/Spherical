// Handles getting the payload.
const span = document.getElementById("payload");
let payload: any;
if (span) {
    payload = JSON.parse(span.innerHTML);
    span.remove();
}
export default payload;
