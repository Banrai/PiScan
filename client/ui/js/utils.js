function showModal (title, header, message) {
    $('#modalWindow').modal('show');
    $('#modalTitle').text(title);
    $('#modalMessageHeader').text(header);
    $('#modalMessage').text(message);
}

function checkAccountStatus (accId) {
    var fn = null;
    if( arguments.length > 1 ) {
	fn = arguments[1];
    }
    $.ajax({type: "POST",
	    url: "/status/",
	    data: { account: accId },
	    dataType: "json",
	    success: function (d) {
		if( !d["msg"] || d["msg"] !== "true" || d["err"] ) {
		    showModal("Warning", "Verification Pending", "Your email address is unverified. Please check your email for the link we sent you.");
		} else {
		    if( fn !== null ) {
			fn();
		    }
		}
	    }
	   });
}
