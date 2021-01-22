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
  mutations.addedNodes.forEach(localizeTimeInNode);
});
timestamper.observe(chatMessages, { childList: true });

// Localize all existing nodes.
chatMessages.childNodes.forEach(localizeTimeInNode);
