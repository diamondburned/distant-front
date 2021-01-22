const chatMessages = document.querySelector("div#chat-box div.chat-messages");

function localizeTimeInNode(node) {
  node.querySelector("time").forEach((time) => {
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
  mutations.addedNodes.forEach((node) => localizeTimeInNode);
});
timestamper.observe(chatMessages, { childList: true });

// Localize all existing nodes.
chatMessages.childNodes.forEach((node) => localizeTimeInNode);
