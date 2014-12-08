$(function(){
    $('a.update').click(function(event){
        event.preventDefault();
        var toggleId = $(this).attr('href');
        $(toggleId).toggle(300);
        $("#accountEmail").focus();
        $(this).toggle();
    });
    $("#accountEmail").focus();
    if( ACC_REG ) {
	var accId = $('input[type=hidden]#account').val(),
	    postData = { account: accId };
	$.ajax({type: "POST",
		url: "/status/",
		data: postData,
                dataType: "json",
		success: function (d) {
		    if( !d["msg"] || d["msg"] !== "true" || d["err"] ) {
			showModal("Warning", "Verification Pending", "Your email address is unverified. Please check your email for the link we sent you.");
		    }
                }
	       });
    }		
});

