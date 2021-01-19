export var Endpoint = "http://45.32.188.77:31878";

async function fetchEndpoint(path, opt = { method: "GET" }) {
  opt.credentials = "same-origin";

  const f = await fetch(`${Endpoint}/${path}`, opt);
  return f.json();
}

// Summary fetches the /summary endpoint.
export async function Summary() {
  return await fetchEndpoint("summary");
}
