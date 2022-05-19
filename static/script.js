window.onload = function () {
    const addressButton = document.getElementById('addressButton');
    const postcardsButton = document.getElementById('postcardsButton');
    const postcardsDiv = document.getElementById('postcardsDiv');
    // const deleteAddressButton = document.getElementById('deleteAddressButton');
    const editAddressButton = document.getElementById('editAddressButton');
    const submitAddress = document.getElementById('submitAddress');
    const addressDiv = document.getElementById('addressDiv')
    const postcardImageInput = document.getElementById("postcardFileInput")
    const cropButton = document.getElementById("crop")
    const submitPreviewPhotoButton = document.getElementById('submitPreviewPhoto');
    const submitPostcardButton = document.getElementById('submitPostcard')
    const submitPhysicalPostcardButton = document.getElementById("submitPhysicalPostcardButton")
    const recipientSelector = document.getElementById('recipientSelector')
    const submitPreviewStatusLabel = document.getElementById('submitPreviewStatusLabel')
    const submitPostcardStatusLabel = document.getElementById('submitPostcardStatusLabel')
    const submitAddressStatusLabel = document.getElementById('submitAddressStatusLabel')
    const backTextArea = document.getElementById("backTextArea")
    const postcardslist = document.getElementById("postcardslist")
    const cannotSendPhysicalPostcardDiv = document.getElementById("cannotSendPhysicalPostcardDiv")
    const physicalPostcardErrorLabel = document.getElementById("physicalPostcardErrorLabel")
    let contactMapping = {};
    let photo;
    let credits = 0;
    let address;

    function updateStripeLink(paymentId, recurseId, email) {
        const queryParams = {
            client_reference_id: recurseId,
            prefilled_email: email
        };
        
        const searchParams = new URLSearchParams(queryParams);
        const stripeLink = document.getElementById("stripeLink");
        stripeLink.href = "https://buy.stripe.com/" + paymentId + "?" + searchParams;
    }

    fetch("/addresses").then(response =>
        response.json()
    ).then(data => {
        address = data;
        document.getElementById("name").innerText = data["name"]
        document.getElementById("address1").innerText = data["address_line1"]
        document.getElementById("address2").innerText = data["address_line2"]
        document.getElementById("city").innerText = data["address_city"]
        document.getElementById("state").innerText = data["address_state"]
        document.getElementById("zip").innerText = data["address_zip"]
        document.getElementById("acceptsPhysicalMail").innerText = data["acceptsPhysicalMail"]

        updateStripeLink(data["stripePaymentLinkId"], data["recurse_id"], data["email"]);
    })

    fetch("/contacts").then(response =>
        response.json()
    ).then(data => {
        credits = data["credits"]
        submitPhysicalPostcardButton.innerText = "Send Physical Postcard ✉️ (" + credits + " credits remaining)"

        let contacts = data["contacts"]
        let rc_opt;
        contacts.forEach(contact => {
            var opt = document.createElement('option')
            rc_id = contact["recurseId"]
            opt.value = rc_id
            var innerText = ""
            if (contact["acceptsPhysicalMail"]) {
                innerText = "📮✅ "
            }
            if(contact["batch"]) {
                innerText += contact["name"] + " (" + contact["batch"] + ")";
            } else {
                innerText += contact["name"];
            }
            opt.innerText = innerText
            if(rc_id === 0) { // for Recurse center
                rc_opt = opt;
            }
            recipientSelector.appendChild(opt)
            contactMapping[contact["recurseId"]] = { 
                "name": contact["name"],
                "acceptsPhysicalMail": contact["acceptsPhysicalMail"]
            };
        });
        rc_opt.selected = true;
        onSelectRecipient();
        return fetch("/postcards");
    }).then(response => response.json()
    ).then(data => {
        postcards = data["data"]
        for (let postcard of postcards) {
            var postcardListItem = document.createElement('li')

            var postcardURL = document.createElement('a')
            postcardURL.href = postcard["url"]
            postcardURL.target = "_blank"
            var postcardDiv = document.createElement('div');
            postcardDiv.classList.add("postcardDiv");

            var timeDiv = document.createElement('div');
            var time = new Date(postcard["date_created"]).toLocaleString();
            timeDiv.innerText = time;

            var from_id = postcard["metadata"]["from_rc_id"]
            let from_name = "Deleted User"
            if (contactMapping[from_id]) {
                from_name = contactMapping[from_id]["name"]
            }

            var senderDiv = document.createElement('div');
            senderDiv.innerText = from_name;

            postcardDiv.appendChild(timeDiv)
            postcardDiv.appendChild(senderDiv)
            postcardURL.appendChild(postcardDiv)
            postcardListItem.appendChild(postcardURL)
            postcardslist.appendChild(postcardListItem)
        }
    });

    $(document).ready(function () {
        $('.js-example-basic-single').select2();
    });

    addressButton.addEventListener('click', function () {
        if (addressDiv.style.display === "none") {
            addressButton.innerText = "Hide my address 👻"
            addressDiv.style.display = "block";
        } else {
            addressButton.innerText = "Show my address :)"
            addressDiv.style.display = "none";
        }
    })

    postcardsButton.addEventListener('click', function () {
        if (postcardsDiv.style.display === "none") {
            postcardsButton.innerText = "Hide my postcards 💌"
            postcardsDiv.style.display = "block";
        } else {
            postcardsButton.innerText = "Show my digital postcards :)"
            postcardsDiv.style.display = "none";
        }
    })

    postcardImageInput.addEventListener('change', function (event) {
        if (event.target.files.length > 0) {
            document.getElementById("postcardImageCropper").style.display = "block";
            const file = event.target.files[0];
            const src = URL.createObjectURL(file);
            const preview = document.getElementById("postcardImagePreview");
            preview.filename = file.name;
            preview.src = src;
            preview.style.display = "block";

            //cropper
            preview.onload = function () {
                const cropper = document.getElementById("cropper");
                const previewRect = preview.getBoundingClientRect();
                if (previewRect.height >= previewRect.width * 4.25 / 6.25) {
                    cropper.style.width = "calc(100% - 20px)";
                    cropper.style.height = `${(previewRect.width - 20) * 4.25 / 6.25}px`;
                } else {
                    cropper.style.height = "calc(100% - 20px)";
                    cropper.style.width = `${(previewRect.height - 20) * 6.25 / 4.25}px`;
                }
                cropper.style.transform = "translate(0px, 0px)";
            }
        }
    })

    let startX, startY;
    let prevTranslateX = 0, prevTranslateY = 0;
    let isRightBottomCornerBeingDragged = false;

    document.addEventListener("dragstart", function (event) {
        if (event.target.id === "cropper") {
            startX = event.clientX;
            startY = event.clientY;
        } else if (event.target.id === "rightBottomCorner") {
            isRightBottomCornerBeingDragged = true;
        }
    }, false);

    /* events fired on the drop targets */
    document.addEventListener("dragover", function (event) {
        // prevent default to allow drop
        event.preventDefault();
    }, false);

    function bound(value, min, max) {
        return Math.min(Math.max(value, min), max);
    }

    document.addEventListener("drop", function (event) {
        // prevent default action (open as link for some elements)
        event.preventDefault();
        const cropper = document.getElementById("cropper");
        const cropperRect = cropper.getBoundingClientRect();
        const preview = document.getElementById("postcardImagePreview");
        const previewRect = preview.getBoundingClientRect();
        if (!isRightBottomCornerBeingDragged) {
            let newX = bound(event.clientX - startX + prevTranslateX, 0, previewRect.width - cropperRect.width);
            let newY = bound(event.clientY - startY + prevTranslateY, 0, previewRect.height - cropperRect.height);
            cropper.style.transform = `translate(${newX}px, ${newY}px)`;
            prevTranslateX = newX;
            prevTranslateY = newY;
        } else {
            let img = new Image();
            img.onload = () => {
                let minWidth = 1875 * previewRect.width / img.naturalWidth;
                let newWidth = Math.max(event.clientX - cropperRect.left, minWidth + 1)
                cropper.style.width = `${newWidth}px`;
                cropper.style.height = `${newWidth * 4.25 / 6.25}px`;
                isRightBottomCornerBeingDragged = false;
            }
            img.src = preview.src;
        }
    }, false);

    cropButton.addEventListener('click', function drawCroppedImageToCanvas() {
        const preview = document.getElementById("postcardImagePreview");
        const cropper = document.getElementById("cropper");
        const previewRect = preview.getBoundingClientRect();
        const cropperRect = cropper.getBoundingClientRect();

        const canvas = document.getElementById('canvas');
        canvas.style.display = "block";
        const ctx = canvas.getContext('2d');
        const img = new Image();
        img.onload = () => {
            canvas.width = cropperRect.width * img.naturalWidth / previewRect.width;
            canvas.height = cropperRect.height * img.naturalHeight / previewRect.height;
            canvas.style.width = `${cropperRect.width}px`;
            canvas.style.height = `${cropperRect.height}px`;

            ctx.drawImage(
                img,
                (cropperRect.left - previewRect.left) * img.naturalWidth / previewRect.width,
                (cropperRect.top - previewRect.top) * img.naturalHeight / previewRect.height,
                cropperRect.width * img.naturalWidth / previewRect.width,
                cropperRect.height * img.naturalHeight / previewRect.height,
                0,
                0,
                cropperRect.width * img.naturalWidth / previewRect.width,
                cropperRect.height * img.naturalHeight / previewRect.height,
            );
            canvas.toBlob((blob) => {
                photo = new File([blob], preview.filename, { type: "image/jpeg" })
                document.getElementById("postcardImageCropper").style.display = "none";
            }, 'image/jpeg');
        };
        img.src = preview.src;
    })

    submitPreviewPhotoButton.addEventListener('click', function () {
        if (!photo) {
            submitPreviewStatusLabel.innerText = "no photo selected"
            submitPreviewStatusLabel.style = "background-color: red"
            return;
        }
        let formData = new FormData()
        formData.append("front-postcard-file", photo)
        var backText = backTextArea.value;
        backText = backText.replace(/( (?= ))/gm, '&nbsp;');
        backText = backText.replace(/(\r\n|\n|\r)/gm, '<br />');
        formData.append("back", backText)
        fetch("/postcards?mode=digital_preview&toRecurseId=0", { method: "POST", body: formData }).then(response =>
            response.json()
        ).then(data => {
            if (!data["err"] && !data["status_code"]) {
                pdfPreviewLink = document.getElementById("pdfPreviewLink")
                pdfPreviewLink.innerText = data['url']
                pdfPreviewLink.href = data['url']
                submitPreviewStatusLabel.style = ""
                submitPreviewStatusLabel.innerText = ""
            } else {
                submitPreviewStatusLabel.innerText = data["message"]
                submitPreviewStatusLabel.style = "background-color: red"
            }
        })
    });

    function onSelectRecipient() {
        let recipientId = recipientSelector.value
        let receipientName = recipientSelector.options[recipientSelector.selectedIndex].innerText
        document.getElementById("sendPostcardHeading").innerText = "Send a postcard to " + receipientName + "!";
        if (contactMapping[recipientId]["acceptsPhysicalMail"] && credits > 0) {
            cannotSendPhysicalPostcardDiv.style.display = "none";
            submitPhysicalPostcardButton.disabled = false
            submitPhysicalPostcardButton.style.backgroundColor = "#0594d8"
        }
        else {
            if (credits <= 0) {
                physicalPostcardErrorLabel.innerText = "No credits available to send physical postcards."
            } else {
                physicalPostcardErrorLabel.innerText = "User is not eligible to recieve physical postcards."
            }
            cannotSendPhysicalPostcardDiv.style.display = "block";
            submitPhysicalPostcardButton.disabled = true
            submitPhysicalPostcardButton.style.backgroundColor = "gray"
        }
    }

    $('.js-example-basic-single').on('select2:select', function (e) {
        onSelectRecipient();
    });

    submitPostcardButton.addEventListener('click', function () {
        let recipientId = recipientSelector.value
        let receipientName = recipientSelector.options[recipientSelector.selectedIndex].innerText
        if (!photo) {
            submitPostcardStatusLabel.innerText = "no photo selected"
            submitPostcardStatusLabel.style = "background-color: red"
            return
        }
        if (!recipientId) {
            submitPostcardStatusLabel.innerText = "no recipient selected"
            return
        }

        let formData = new FormData()
        formData.append("front-postcard-file", photo)
        formData.append("back", backTextArea.value)
        fetch("/postcards?mode=digital_send&toRecurseId=" + recipientId, { method: "POST", body: formData }).then(response =>
            response.json()
        ).then(data => {
            if (!data["err"] && !data["status_code"]) {
                submitPostcardStatusLabel.innerText = "success sending to " + receipientName + " ✅"
                submitPostcardStatusLabel.style = "background-color: green"
            } else {
                submitPostcardStatusLabel.innerText = data["message"]
                submitPostcardStatusLabel.style = "background-color: red"
            }
        })
    })

    submitPhysicalPostcardButton.addEventListener('click', function () {
        let recipientId = recipientSelector.value
        let receipientName = recipientSelector.options[recipientSelector.selectedIndex].innerText
        if (!photo) {
            submitPostcardStatusLabel.innerText = "no photo selected"
            submitPostcardStatusLabel.style = "background-color: red"
            return
        }
        if (!recipientId) {
            submitPostcardStatusLabel.innerText = "no recipient selected"
            return
        }

        let formData = new FormData()
        formData.append("front-postcard-file", photo)
        formData.append("back", backTextArea.value)
        fetch("/postcards?mode=physical_send&toRecurseId=" + recipientId, { method: "POST", body: formData }).then(response =>
            response.json()
        ).then(data => {
            if (!data["err"] && !data["status_code"]) {
                credits = data["credits"]
                submitPostcardStatusLabel.innerText = "success sending physical mail to " + receipientName + " ✅. Credits remaining: " + credits
                submitPostcardStatusLabel.style = "background-color: green"
                submitPhysicalPostcardButton.innerText = "Send Physical Postcard ✉️ (" + credits + " credits remaining)"
            } else {
                submitPostcardStatusLabel.innerText = data["message"]
                submitPostcardStatusLabel.style = "background-color: red"
            }
        }).catch(function (error) {
            submitPostcardStatusLabel.innerText = "Error sending postcard."
            submitPostcardStatusLabel.style = "background-color: red"
        })

    })

    editAddressButton.addEventListener('click', function () {
        if (document.getElementById('editAddressDiv').style.display === "none") {
            document.getElementById('showAddressDiv').style.display = "none";
            document.getElementById('editAddressDiv').style.display = "block";

            editAddressButton.innerText = "Cancel editing";

            document.getElementById("editName").value = address["name"]
            document.getElementById("editAddress1").value = address["address_line1"]
            document.getElementById("editAddress2").value = address["address_line2"]
            document.getElementById("editCity").value = address["address_city"]
            document.getElementById("editState").value = address["address_state"]
            document.getElementById("editZip").value = address["address_zip"]
            document.getElementById("editReceivePhysicalMail").checked = address["acceptsPhysicalMail"]
        } else {
            document.getElementById('showAddressDiv').style.display = "block";
            document.getElementById('editAddressDiv').style.display = "none";
            editAddressButton.innerText = "Edit my address";
        }
    })

    submitAddress.addEventListener('click', function () {
        address = {
            "name": document.getElementById("editName").value,
            "address1": document.getElementById("editAddress1").value,
            "address2": document.getElementById("editAddress2").value,
            "city": document.getElementById("editCity").value,
            "state": document.getElementById("editState").value,
            "zip": document.getElementById("editZip").value,
            "acceptsPhysicalMail": document.getElementById("editReceivePhysicalMail").checked
        }
        body = new URLSearchParams(address)

        fetch("/addresses", { method: "POST", body: body }).then(response => {
            if (response.status == 200) {
                return response.json().then(data => {
                    window.location.replace(window.location.href);
                });
            } else {
                submitAddressStatusLabel.style = "background-color: red"
                submitAddressStatusLabel.innerText = "Invalid address, please try again!"
            }
        })
    });

    // deleteAddressButton.addEventListener('click', function () {
    //     fetch("/addresses", { method: "DELETE" }).then(response => {
    //         window.location.replace(window.location.href);
    //         return;
    //     });
    // })
}
