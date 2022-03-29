// Oauth request function
async function reqOauth() {
    let response = await request("../auth/google/login", "GET");
    console.log(response);
}

// Generic request function
async function request(endpoint, requestMethod/*, requestBody*/) {
    let response = await fetch(endpoint, {
        method: requestMethod,
        header: {
            "Content-Type": "application/json",
            "charset": "UTF-8"
        }//,
        //body: JSON.stringify(requestBody)
    });
    let returnValueText = await response.json();
    return returnValueText;
}

async function oauthLogin() {
    //get request to /auth/google/login and let the server redirect to the google login page
    let resp = await fetch("/auth/google/login");
    //follow the redirect to the google login page
    let j = await resp.json();
    if (j.error) {
        alert(j.msg);
    }
    window.location.href = j.data.url;
}