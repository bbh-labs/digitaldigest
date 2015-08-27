function signInCallback(authResult) {
	if (authResult["code"]) {
		// Send the code to the server
		$.ajax({
			type: "POST",
			url: "/login",
			data: {authCode: authResult["code"]},
		}).done(function(resp) {
			window.location.reload();
		}).fail(function(resp) {
			Materialize.toast(resp.statusText, 3000, "red white-text");
		});
	} else {
		// There was an error.
		console.log("There was an error!");
	}
}

function onPlatformReady() {
	gapi.load("auth2", function() {
		auth2 = gapi.auth2.init({
			client_id: "275859936684-90o26gr4hdbr4jgvdjobuath4qhq90fc.apps.googleusercontent.com",
		});
	});
}

function signIn() {
	auth2.grantOfflineAccess({"redirect_uri": "postmessage"}).then(signInCallback);
}

function signOut() {
	$.ajax({
		type: "POST",
		url: "/logout",
	}).done(function(resp) {
		if (auth2) auth2.signOut();
		if (gapi && gapi.auth) gapi.auth.setToken(null);
		window.location.reload();
	});
}
