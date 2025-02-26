// src/pages/_app.tsx
import { AppProps } from "next/app";
import "@/styles/globals.css";
import Head from "next/head";

function MyApp({ Component, pageProps }: AppProps) {
  return (
    <>
      <Head>
        <title>Identity Graph Visualization</title>
        <meta
          name="description"
          content="Explore identity relationships and permissions"
        />
        <link rel="icon" href="/favicon.ico" />
      </Head>
      <Component {...pageProps} />
    </>
  );
}

export default MyApp;
