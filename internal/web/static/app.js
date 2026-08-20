async function fetchJSON(url, opts) {
  const res = await fetch(url, opts);
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

function fmtTime(ts) {
  if (!ts) return "—";
  return new Date(ts).toLocaleString("zh-CN");
}

function renderSpikes(summaries) {
  const el = document.getElementById("spike-table");
  if (!summaries.length) {
    el.textContent = "暂无尖峰事件";
    return;
  }
  const rows = summaries.flatMap((s) =>
    (s.recent || []).slice(-5).map((ev) => `
      <tr>
        <td>${ev.belt_id}</td>
        <td>${ev.peak.toFixed(2)} g</td>
        <td>${fmtTime(ev.start)}</td>
        <td>${fmtTime(ev.end)}</td>
        <td>${ev.sample_count}</td>
      </tr>
    `)
  );
  el.innerHTML = `
    <table>
      <thead>
        <tr>
          <th>皮带</th>
          <th>峰值</th>
          <th>开始</th>
          <th>结束</th>
          <th>样本数</th>
        </tr>
      </thead>
      <tbody>${rows.join("")}</tbody>
    </table>
  `;
}

function renderBelts(belts) {
  const el = document.getElementById("belt-stats");
  if (!belts.length) {
    el.textContent = "暂无缓冲数据";
    return;
  }
  el.innerHTML = belts.map((b) => `
    <div class="card">
      <div><strong>皮带 ${b.belt_id}</strong></div>
      <div>缓冲 ${b.count} / ${b.capacity}</div>
      <div>最新 ${fmtTime(b.newest_ts)}</div>
    </div>
  `).join("");
}

async function refresh() {
  try {
    const stats = await fetchJSON("/v1/stats");
    renderSpikes(stats.spikes || []);
    renderBelts(stats.belts || []);
  } catch (err) {
    console.error(err);
  }
}

document.getElementById("sample-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const fd = new FormData(e.target);
  const payload = {
    belt_id: fd.get("belt_id"),
    accel: Number(fd.get("accel")),
    ts: new Date().toISOString(),
  };
  try {
    const out = await fetchJSON("/v1/samples", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    document.getElementById("sample-result").textContent = JSON.stringify(out, null, 2);
    refresh();
  } catch (err) {
    document.getElementById("sample-result").textContent = String(err);
  }
});

refresh();
setInterval(refresh, 5000);
