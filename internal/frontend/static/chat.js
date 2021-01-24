const chatMessages = document.querySelector("#chat-box > .chat-messages");
const timeFmt = new Intl.DateTimeFormat(undefined, { timeStyle: "short" });

function localizeTimeInNode(node) {
  // Exit if the current node is not an Element.
  if (!node.querySelectorAll) return;

  node.querySelectorAll("time").forEach((time) => {
    try {
      const date = new Date(time.getAttribute("datetime"));
      time.innerText = timeFmt.format(date);
    } catch (err) {
      // best effort
      console.error(`failed to localize timestamp: ${err}`);
    }
  });
}

// Localize new nodes.
const timestamper = new MutationObserver((mutations) => {
  mutations.forEach((mut) => mut.addedNodes.forEach(localizeTimeInNode));
});
timestamper.observe(chatMessages, { childList: true });

// Localize all existing nodes.
chatMessages.childNodes.forEach(localizeTimeInNode);

function hasCookies(...cookies) {
  var found = 0;

  return document.cookie.split("; ").some((it) => {
    const thisName = it.trim();

    const ok = cookies.some((name) => thisName.startsWith(`${name}=`));
    if (ok) found++;

    return found === cookies.length;
  });
}

// LastSelector is the selector for the last message. Since we're doing
// column-reverse, the last message is the first one in the DOM tree.
const LastSelector = "div.chat-messages > div.chat-message:first-child";

// listen opens a persistent HTTP connection to receive null-delimited HTML
// chunks.
async function listen() {
  const last = document.querySelector(LastSelector);
  const resp = await fetch(`/chat/listen/${last ? last.id : ""}`);
  const utf8 = new TextDecoder("utf-8");

  const reader = resp.body.getReader();
  var textBuf = [];

  let packet = {};

  while (true) {
    const packet = await reader.read();
    if (!packet || packet.done) throw "stream closed";
    if (!packet.value) continue; // empty read; not an error until EOF.

    let text = utf8.decode(packet.value);
    console.log("received", text);

    // Iterate until we're out of delimiters.
    while (true) {
      let delim = text.indexOf("\0");
      if (delim < 0) {
        // Buffer, stop looking and continue reading.
        textBuf.push(text);
        break;
      }

      // Push the complete segment and join.
      textBuf.push(text.slice(0, delim));
      const html = textBuf.join("");

      // Clear the buffer but allow space reusing and add the tail of the chunk
      // in.
      textBuf.length = 0;
      text = text.slice(delim + 1);

      // Write the received HTML chunk.
      addMessageHTML(html);
    }
  }
}

function addMessageHTML(html) {
  if (!html) return;

  chatMessages.insertAdjacentHTML("afterbegin", html);

  // Clean up messages. This preserves the last 50 messages.
  const elems = chatMessages.getElementsByClassName("chat-message");
  for (let i = 50; i < elems.length; i++) {
    chatMessages.removeChild(elems[i]);
  }
}

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// backgroundLoop is the long polling loop.
async function backgroundLoop() {
  while (true) {
    console.log("listening...");

    // Cooldown for 5 seconds minimum before each retry. Using allSettled
    // allows us to sleep even if listen has thrown an exception.
    const [res, _] = await Promise.allSettled([listen(), sleep(10000)]);

    if (res.status === "rejected") {
      console.error(`Listen error: ${res.reason}`);
    }
  }
}

// Start the background loop.
backgroundLoop();

const chatSend = document.querySelector("form#chat-send"),
  chatInput = chatSend.querySelector("input[type='text']"),
  chatButton = chatSend.querySelector("button[type='submit']");

// Start binding the sending form to remove the need to reload the page.

chatInput.addEventListener("keydown", async (ev) => {
  if (ev.key === "Enter") {
    ev.preventDefault();
    await sendMessage();
  }
});

chatButton.addEventListener("click", async (ev) => {
  ev.preventDefault();
  await sendMessage();
});

async function sendMessage() {
  const m = chatInput.value;
  if (!m) return;

  chatButton.disabled = true;
  chatInput.disabled = true;
  chatInput.value = "";

  try {
    const r = await fetch(`/chat?m=${encodeURIComponent(m)}`, {
      method: "POST",
      redirect: "manual",
      credentials: "same-origin",
    });
    // Expect a redirection on success.
    if (r.type != "opaqueredirect") {
      throw `unexpected ${r.status} response: ${await r.text()}`;
    }
  } catch (err) {
    chatInput.value = m;
    console.error(`failed to send message: ${err}`);
  }

  chatButton.disabled = false;
  chatInput.disabled = false;
}
