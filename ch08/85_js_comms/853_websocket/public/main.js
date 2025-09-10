const log = document.getElementById("log");
const ws = new WebSocket("ws://localhost:18888/ws");

ws.onopen = () => {
  log.textContent += "✅ 接続しました\n";
};

ws.onmessage = (event) => {
  log.textContent += "📩 " + event.data + "」\n";
};

document.getElementById("send").addEventListener("click", () => {
  const input = document.getElementById("msg");
  ws.send(input.value);
  log.textContent += "➡️ " + input.value + "\n";
  input.value = "";
});
