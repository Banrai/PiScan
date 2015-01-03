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

function showConfirmModal (title, header, message, continueAction) {
    var continueButton = null,
	dismissButton  = null;
    if( arguments.length > 4 ) {
	continueButton = arguments[4];
    }
    if( arguments.length > 5 ) {
	dismissButton = arguments[5];
    }
    
    $('#modalWindow').modal('show');
    $('#modalTitle').text(title);
    $('#modalMessageHeader').text(header);
    $('#modalMessage').text(message);
    $('#modalContinue').show();
    $('#modalContinue').attr('href', continueAction); 
    if( continueButton !== null ) {
	$('#modalContinueButton').text(continueButton);
    }
    if( dismissButton !== null ) {
	$('#modalDismissButton').text(dismissButton);
    }
}

var confirmShutdown = function(event){
    event.preventDefault();
    showConfirmModal('Shutdown', 'Shutdown this scanner', 'Are you sure you want to shutdown?', $(this).attr('href'), 'Yes', 'Cancel');
};
