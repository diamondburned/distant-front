import React from "https://esm.sh/react";
import { Head } from "https://deno.land/x/aleph/mod.ts";

import Header from "./components/header.js";

import "https://unpkg.com/spectre.css/dist/spectre.min.css";
import "https://unpkg.com/spectre.css/dist/spectre-exp.min.css";
import "https://unpkg.com/spectre.css/dist/spectre-icons.min.css";

export default function App({ Page, pageProps }) {
  return (
    <>
      <Head>
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>Distant Front - Distance Server</title>
      </Head>
      <Header />
      <Page {...pageProps} />
    </>
  );
}
