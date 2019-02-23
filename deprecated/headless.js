const puppeteer = require('puppeteer');
const fs = require('fs');
var request = require('request');

(async () => {
  target = "http://localhost:3000"
  // Grab routes dump
  const TARGET_APP_PATH = process.env.TARGET_APP_PATH;
  // Const routes_path = "/tmp/debug";
  const routes_path = TARGET_APP_PATH + "/.meta/" + "final_routes.json";
  const routes_file = fs.readFileSync(routes_path, 'utf8');
  routes = JSON.parse(routes_file);
  // Set cookie
  const cookie = fs.readFileSync("./cookie", 'utf8')
  // Init browser
  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  // TODO: On alert raise exception
  page.on('dialog', msg => {
    console.log(msg);
  });
  // Iterate through routes
  for (route of routes) {
      await new Promise((resolve, reject) => {
        route["segments"].forEach(segment => {
           route["path"] = route["path"].replace(':'+ segment, route["params"][segment])
        });
        request({
            'url' : target + route['path'],
            'method' : route['verb'],
            'headers' : {'Cookie' : cookie.trim()},
            'json' : route["params"]
        },
        async (err, resp, body) => {
            if (resp.headers['content-type'].includes('text/html')) {
                await page.setContent(body)
                await page.screenshot({path: 'daniel/test_' + route["path"].replace(/\//g,'') + ".png"});
            }
            resolve()
        });
      });
  };
  await browser.close();
})();