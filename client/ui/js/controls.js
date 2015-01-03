
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

function buyAmzUS (targetClass) {
    var i, d, q; cartUrl = "", amzItems = {}, selectedItems = [];
    $("input:hidden").each(function() {
	if( $(this).attr("class") == targetClass ) {
	    amzItems[ $(this).attr("name") ] = $(this).attr("value");
	}
    });
    $(".chk_item").each(function() {
	if( $(this).is(":checked") ) {
	    var val = $(this).attr("value"),
	       asin = amzItems[val];
	    if( asin ) {
		selectedItems.push(asin);
	    }
	}
    });
    for(i=0, d=selectedItems.length; i<d; i++) {
	q = i+1;
	if( i>0 ) { cartUrl += "&"; }
	cartUrl += "ASIN."+q+"="+selectedItems[i]+"&Quantity."+q+"=1";
    }
    if( cartUrl.length > 0 ) {
	location.href = "http://www.amazon.com/gp/aws/cart/add.html?" + cartUrl;
    } else {
	// none of the selected items can be bought from amzUS
	showModal("Sorry", "Unavailable", "None of the selected items can be purchased from Amazon (US)");
    }
}

$(function(){
    if( ! Modernizr.canvas || ! Modernizr.svg ) {
	window.location.href = '/browser';
    }
    $('a.shutdown').click(confirmShutdown);
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
			showModal("Sorry", "Error", "That item could not be deleted");
		    }
                },
                error: function (d) {
		    if( d["err"] ) {
			showModal("Sorry", "Error", d[err]);
		    } else {
			showModal("Sorry", "Error", "There was a problem deleting that item");
		    }
		}
	       });
    });
    $(".dropdown-menu li a").on("click", function(event) {
	event.preventDefault();
	if( anyItemChecked() ) {
	    var target = $(this).attr('href');
	    if( "/buyAMZN:us/" == target ) {
		buyAmzUS("AMZN:us");
	    }  else {
		var submitOk = function () {
		    $('#bulkActions').attr('action', target);
		    $("#bulkActions").submit();
		};
		if( "/email/" == target ) {
		    checkAccountStatus( $('input[type=hidden]#account').val(), submitOk );
		} else {
		    submitOk();
		}
	    }
	}
    });
});

