function signInCallback(authResult) {
	if (authResult["code"]) {
		// Send the code to the server
		$.ajax({
			type: "POST",
			url: "/login",
			success: function(result) {
				window.location.reload();
			},
			data: {authCode: authResult["code"]},
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
