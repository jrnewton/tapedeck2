This code is from `stream-archive.tsx` and defines how the list is built on https://whrb.org/stream-archive/. 

 __Note__ that times are in _UTC_.
```
import * as React from "react";
import { useRef, useEffect, useState } from "react";
import { PageProps } from "gatsby";
import ContentLayout from "../components/contentLayout";
import Hls from "hls.js";
import SEO from "../components/seo";
// import ArchivePlayer from "../components/archivePlayer";
import ArchiveList from "../components/archiveList";

export const Head = () => (
  <SEO title="Stream Archive" description="Archive of recent shows on WHRB" />
);

const hls = new Hls();

const ARCHIVE_BASE_URL = "https://stream.whrb.org/archive";

const formatter = new Intl.DateTimeFormat("en-US", {
  day: "2-digit",
  month: "2-digit",
  year: "numeric",
  hour: "2-digit",
  hour12: false,
  timeZone: "UTC",
});
const readableFormatter = new Intl.DateTimeFormat("en-US", {
  day: "numeric",
  weekday: "short",
  month: "short",
  hour: "numeric",
  timeZoneName: "short",
});

const StreamArchivePage: React.FC<PageProps> = () => {
  const audioRef = useRef<HTMLAudioElement>(null);
  const [playing, setPlaying] = useState("");

  const startAudio = (audioFile: string) => {
    hls.loadSource(`${ARCHIVE_BASE_URL}/${audioFile}/${audioFile}.m3u8`);
    if (audioRef.current) {
      audioRef.current.play();
    }
  };

  useEffect(() => {
    if (audioRef.current) {
      hls.attachMedia(audioRef.current);
    }
  }, []);

  const dateParts: string[][] = [];

  for (let curHour = 1; curHour <= 48; curHour += 1) {
    const d = new Date();
    d.setHours(d.getHours() - curHour);
    const parts = formatter.formatToParts(d);
    const year = parts.find((p) => p.type === "year");
    const month = parts.find((p) => p.type === "month");
    const day = parts.find((p) => p.type === "day");
    const hour = parts.find((p) => p.type === "hour");

    const filename = `${year?.value}_${month?.value}_${day?.value}_${hour?.value}`;
    const readable = readableFormatter.format(d);
    dateParts.push([filename, readable]);
  }

  const getStartHandler = (filename: string, readableName: string) => {
    return () => {
      startAudio(filename);
      setPlaying(readableName);
    };
  };

  return (
    <ContentLayout pageClass="page-type-page">
      <section className="main-content">
        <h1>WHRB Stream Archive</h1>
        <p><em>For legal reasons, archives of programs are available for only two days</em></p>
        {playing ? <p>Playing: {playing}</p> : <p>Click a time to listen</p>}
        <audio ref={audioRef}></audio>
        {playing && audioRef.current && (audioRef.current.controls = true)}
        <audio ref={audioRef} style={{ width: "100%", marginBottom: "20px" }}></audio>
        <ArchiveList dateParts={dateParts} getStartHandler={getStartHandler} />
        </section>
    </ContentLayout>
  );
};
export default StreamArchivePage;
```