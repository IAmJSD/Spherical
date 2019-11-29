// Handles getting the payload.
const span = document.getElementById("payload")
let payload: any
if (span) {
    payload = JSON.parse(span.innerText)
    span.remove()
}
export default payload
