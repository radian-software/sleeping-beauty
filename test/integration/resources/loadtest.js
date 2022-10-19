import http from "k6/http";
import { sleep } from "k6";

export const options = {
  vus: 1000,
  duration: "15s",
  thresholds: {
    http_req_failed: ["rate<=0"],
  },
};

export default function () {
  http.get("http://localhost:4444/about");
  sleep(1);
}
