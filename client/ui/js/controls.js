
function anyItemChecked () {
    var checkedBoxes = $(".chk_item:checkbox:checked");
    return (checkedBoxes.length > 0);
}

function toggleActions () {
    if( anyItemChecked() ) {
	$("#id_actions").show();
    } else {
	$("#id_actions").hide();
	$("#id_actions_chk").prop('checked', false);
    }
}

$(function(){
    $("#id_actions_chk").on("click", function() {
	var state = $(this).is(':checked');
	$(".chk_item").each(function() {
	    $(this).prop('checked', state);
	});
	toggleActions();
    });
    $(".chk_item").on("click", function() {
	toggleActions();
    });
});

