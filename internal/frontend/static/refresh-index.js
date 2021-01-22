const mainSelector = document.getElementById("main");
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
  template: (className) => `
	<div class="refresh-bar ${className}">
	  <div class="refresh-inner"></div>
	</div>`,
  allowHTML: true,
});

async function update() {
  loading.data = "refreshing";
  loading.render();

  try {
    const resp = await fetch("");
    main.data = resp.text();
    main.render();

    loading.data = "waiting";
  } catch (err) {
    loading.data = "error";
  }

  loading.render();
}

setInterval(update, 5000);
loading.render();