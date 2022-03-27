// Generic request function
async function request (endpoint, requestMethod, requestBody) {
    let response = await fetch(endpoint, {
        method: requestMethod,
        header: {
            "Content-Type": "application/json",
			"charset": "UTF-8"
        },
        body: JSON.stringify(requestBody)
    });
    let returnValueText = await response.json();
    return returnValueText;
}