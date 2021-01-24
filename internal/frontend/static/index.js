function changeChatPopup() {
  document.querySelectorAll("a[href='/chat']").forEach((a) => {
    a.href = "#";
    a.addEventListener("click", () => {
      window.open("/chat", "popup", "width=450, height=320");
    });
  });
}

const mainSelector = document.querySelector(".refresh-me");
const main = new Reef(mainSelector, {
  data: "",
  template: (data) => data,
  allowHTML: true,
});

const loadingSelector = document.createElement("section");
loadingSelector.classList.add("navbar-section");
loadingSelector.classList.add("loading-indicator");
document.querySelector("header.navbar").appendChild(loadingSelector);

const loading = new Reef(loadingSelector, {
  data: "waiting",
  template: (className) =>
    `
	<div class="refresh-bar ${className}">
	  <div class="refresh-inner"></div>
	</div>`,
  allowHTML: true,
});

async function update() {
  loading.data = "refreshing";
  loading.render();

  try {
    const resp = await fetch("body");
    main.data = await resp.text();
    main.render();
    changeChatPopup();

    loading.data = "waiting";
  } catch (err) {
    loading.data = "error";
  }

  loading.render();
}

setInterval(update, 3500);
loading.render();
changeChatPopup();
