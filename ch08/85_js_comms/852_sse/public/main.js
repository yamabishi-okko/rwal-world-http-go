document.getElementById("start").addEventListener("click", () => {
    const log = document.getElementById("log");
  
    // サーバーの /events に接続
    const evtSource = new EventSource("/events");
  
    evtSource.onmessage = (event) => {
      log.textContent += event.data + "\n";
    };
  
    evtSource.onerror = (err) => {
      log.textContent += "エラー: " + err + "\n";
      evtSource.close();
    };
  });
  