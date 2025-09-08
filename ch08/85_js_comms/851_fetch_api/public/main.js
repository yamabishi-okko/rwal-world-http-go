const out = document.getElementById('out');

document.getElementById('get').onclick = async () => {
  const r = await fetch('/api/hello');
  out.textContent = JSON.stringify(await r.json(), null, 2);
};

document.getElementById('stream').onclick = async () => {
  const r = await fetch('/api/stream');
  const reader = r.body.getReader();
  const dec = new TextDecoder();
  out.textContent = '';
  while (true) {
    const {value, done} = await reader.read();
    if (done) break;
    out.textContent += dec.decode(value);
  }
};

document.getElementById('download').onclick = () => {
  window.location.href = '/api/download';
};

document.getElementById('up').onsubmit = async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const r = await fetch('/api/upload', { method: 'POST', body: fd });
  out.textContent = JSON.stringify(await r.json(), null, 2);
};
