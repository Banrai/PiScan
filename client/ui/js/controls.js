
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
    $('a.trash').on("click", function (event) {
	event.preventDefault();
	var itemId = $(this).attr('href').split('#')[1],
	  postData = { itemId: itemId };
	$.ajax({type: "POST",
		url: "/remove/",
                data: postData,
                dataType: "json",
                success: function (d) {
		    if( d["msg"] && d["msg"] == "Ok" && !d["err"] ) {
			$("#Item_"+itemId).remove();
		    } else {
			alert("Sorry, that item could not be deleted");
		    }
                },
                error: function (d) {
		    if( d["err"] ) {
			alert(d["err"])
		    } else {
			alert("Sorry, there was a problem deleting that item");
		    }
		}
	       });
    });
    $(".dropdown-menu li a").on("click", function(event) {
	event.preventDefault();
	if( anyItemChecked() ) {
	    var target = $(this).attr('href');
	    $('#bulkActions').attr('action', target);
	    $("#bulkActions").submit();
	}
    });
});

