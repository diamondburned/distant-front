const chatMessages = document.querySelector("#chat-box > .chat-messages");

function localizeTimeInNode(node) {
  // Exit if the current node is not an Element.
  if (!node.querySelectorAll) return;

  node.querySelectorAll("time").forEach((time) => {
    try {
      const date = new Date(time.getAttribute("datetime"));
      time.innerText = date.toLocaleTimeString();
    } catch (err) {
      // best effort
      console.error(`failed to localize timestamp: ${err}`);
    }
  });
}

// Localize new nodes.
const timestamper = new MutationObserver((mutations) => {
  if (mutations) mutations.addedNodes.forEach(localizeTimeInNode);
});
timestamper.observe(chatMessages, { childList: true });

// Localize all existing nodes.
chatMessages.childNodes.forEach(localizeTimeInNode);

// MinWait describes the constant (in milliseconds) for the minimum time between
// server polls.
const MinWait = 500;

// LastSelector is the selector for the last message. Since we're doing
// column-reverse, the last message is the first one in the DOM tree.
const LastSelector = "div.chat-messages > div.chat-message:first-child";

// MaxBackoff is the number to slowly increment in the case of error.
const MaxBackoff = 5;

var backoff = 0;
var backingOff = false;

async function poll() {
  if (backoff == MaxBackoff || backingOff) {
    backoff--;
    backingOff = backoff != 0;

    console.log("backing off...");
    return;
  }

  const last = document.querySelector(LastSelector);

  try {
    const resp = await fetch(`after/${last ? last.id : ""}`);
    const html = await resp.text();

    if (html) chatMessages.insertAdjacentHTML("afterbegin", html);
    backoff = 0;

    // Clean up messages.
    if (chatMessages.childNodes.length > 50) {
      for (let i = 50; i < chatMessages.childNodes.length; i++) {
        chatMessages.removeChild(chatMessages.childNodes[i]);
      }
    }
  } catch (err) {
    console.error(`failed to fetch messages: ${err}`);
    backoff++;
  }
}

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// backgroundLoop is the long polling loop.
async function backgroundLoop() {
  while (true) {
    // Wait for either poll() to be done or the minimum waiting duration to be
    // done.
    await Promise.all([poll(), sleep(MinWait)]);
  }
}

// Start the background loop.
backgroundLoop();
