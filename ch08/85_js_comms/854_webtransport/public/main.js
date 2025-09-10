async function connect() {
    log("🔌 WebTransportを初期化中…");
  
    try {
      const transport = new WebTransport("https://localhost:4433/echo");
      await transport.ready;
      log("✅ 接続成功 (実際にはサーバ未実装)");
  
      const writer = transport.datagrams.writable.getWriter();
      const reader = transport.datagrams.readable.getReader();
  
      // 送信
      writer.write(new TextEncoder().encode("hello WebTransport!"));
  
      // 受信ループ（未実装なので動作しない想定）
      while (true) {
        const { value, done } = await reader.read();
        if (done) break;
        log("📩 " + new TextDecoder().decode(value));
      }
    } catch (e) {
      log("❌ エラー: " + e);
    }
  }
  
  function log(msg) {
    const pre = document.getElementById("log");
    pre.textContent += msg + "\n";
  }
  