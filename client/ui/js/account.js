$(function(){
    $('a.shutdown').click(confirmShutdown);
    $('a.update').click(function(event){
        event.preventDefault();
        var toggleId = $(this).attr('href');
        $(toggleId).toggle(300);
        $("#accountEmail").focus();
        $(this).toggle();
    });
    $("#accountEmail").focus();
    if( ACC_REG ) {
	checkAccountStatus( $('input[type=hidden]#account').val() );
    }		
});

