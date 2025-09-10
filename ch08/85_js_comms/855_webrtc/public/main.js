const $ = sel => document.querySelector(sel);
const chat = $("#log");

function scrollBottom() { chat.scrollTop = chat.scrollHeight; }
function addBubble(text, role = "sys", extraClass = "") {
  const p = document.createElement("div");
  p.className = `bubble ${role} ${extraClass}`.trim();
  p.textContent = text;
  chat.appendChild(p);
  scrollBottom();
}

let ws;                 // signaling (WebSocket)
let pc;                 // RTCPeerConnection
let dc;                 // RTCDataChannel
let started = false;
let room = "lobby";

const iceServers = [{ urls: "stun:stun.l.google.com:19302" }];

$("#connect").onclick = () => connectWS();
$("#start").onclick   = () => startOfferFlow();
$("#send").onclick    = () => sendMessage();
$("#msg").addEventListener("keydown", e => { if (e.key === "Enter") sendMessage(); });

function setStatus(s) { $("#status").textContent = s; }

function connectWS() {
  if (ws && ws.readyState === WebSocket.OPEN) return;
  room = $("#room").value || "lobby";
  ws = new WebSocket(`ws://${location.hostname}:18891/ws?room=${encodeURIComponent(room)}`);

  ws.onopen    = () => { addBubble("✅ シグナリング接続しました", "sys"); setStatus("🟢 signaling connected"); };
  ws.onclose   = () => { addBubble("⏹ シグナリング切断されました", "sys"); setStatus("⏸ disconnected"); };
  ws.onerror   = () => { addBubble("⚠️ シグナリングエラー", "sys"); };

  ws.onmessage = async (ev) => {
    const msg = JSON.parse(ev.data);
    if (msg.type === "offer") {
      await ensurePC();
      await pc.setRemoteDescription(msg.sdp);
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);
      ws.send(JSON.stringify({ type: "answer", sdp: pc.localDescription }));
      addBubble("📩 オファー受信 → アンサー送信", "sys");
    } else if (msg.type === "answer") {
      await pc.setRemoteDescription(msg.sdp);
      addBubble("📩 アンサー受信", "sys");
    } else if (msg.type === "candidate") {
      if (msg.candidate) {
        try { await pc.addIceCandidate(msg.candidate); } catch {}
      }
    }
  };
}

async function ensurePC() {
  if (pc) return;
  pc = new RTCPeerConnection({ iceServers });
  pc.onicecandidate = (e) => {
    if (e.candidate && ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: "candidate", candidate: e.candidate }));
    }
  };
  pc.onconnectionstatechange = () => setStatus("🔗 pc: " + pc.connectionState);
  pc.ondatachannel = (ev) => { dc = ev.channel; wireDataChannel(); };
}

function wireDataChannel() {
    dc.onopen  = () => addBubble("🔗 データチャネル開通", "sys");
    dc.onclose = () => addBubble("⏹ データチャネル閉鎖", "sys");
  dc.onmessage = (ev) => {
    const text = String(ev.data);

    // 相手の吹き出し（黄緑）
    if (text === "やりますねぇ！") {
      addBubble(text, "peer", "big"); // 特大表示
    } else {
      addBubble(text, "peer");
    }

    // 「やりますか？」に自動返信する遊び心
    if (text === "やりますか？" && dc.readyState === "open") {
      dc.send("やりますねぇ！");
      addBubble("やりますねぇ！", "me", "big"); // 自分側にも特大表示
    }
  };
}

async function startOfferFlow() {
  await connectWSIfNeeded();
  await ensurePC();
  if (started) return;

  // 自分＝Offer側：DataChannel作成
  dc = pc.createDataChannel("chat");
  wireDataChannel();

  const offer = await pc.createOffer();
  await pc.setLocalDescription(offer);
  ws.send(JSON.stringify({ type: "offer", sdp: pc.localDescription }));
  addBubble("offer sent", "sys");
  started = true;
}

async function connectWSIfNeeded() {
  if (!ws || ws.readyState !== WebSocket.OPEN) {
    connectWS();
    await new Promise(res => {
      const t = setInterval(() => {
        if (ws && ws.readyState === WebSocket.OPEN) { clearInterval(t); res(); }
      }, 50);
    });
  }
}

function sendMessage() {
  const text = $("#msg").value.trim();
  if (!text) return;
  if (!dc || dc.readyState !== "open") {
    addBubble("DataChannel not open yet", "sys");
    return;
  }

  // 自分の吹き出し（⽔⾊）
  if (text === "やりますねぇ！") {
    addBubble(text, "me", "big");
  } else {
    addBubble(text, "me");
  }

  dc.send(text);
  $("#msg").value = "";
}
