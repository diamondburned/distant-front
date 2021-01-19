import React, { useEffect, useState } from "https://esm.sh/react";

import * as api from "../lib/api.js";

export default function Home() {
  const [summary, setSummary] = useState({
    ChatLog: [],
  });
  const update = async () => setSummary(await api.Summary());

  useEffect(async () => {
    await update();
    let iv = setInterval(update, 2000);
    return () => clearInterval(iv);
  }, []);

  let chatMessages = summary.ChatLog.map((msg) => (
    <div className="message">
      <span>{msg.Sender}</span>
      <p>{msg.Chat}</p>
    </div>
  ));

  return (
    <main className="index">
      <section className="messages">{chatMessages}</section>
    </main>
  );
}
