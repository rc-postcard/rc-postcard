window.onload = function () {
    const addressButton = document.getElementById('addressButton');
    const postcardsButton = document.getElementById('postcardsButton');
    const postcardsDiv = document.getElementById('postcardsDiv');
    const deleteAddressButton = document.getElementById('deleteAddressButton');
    const addressDiv = document.getElementById('addressDiv')
    const postcardImageInput = document.getElementById("postcardFileInput")
    const cropButton = document.getElementById("crop")
    const submitPreviewPhotoButton = document.getElementById('submitPreviewPhoto');
    const submitPostcardButton = document.getElementById('submitPostcard')
    const recipientSelector = document.getElementById('recipientSelector')
    const submitPreviewStatusLabel = document.getElementById('submitPreviewStatusLabel')
    const submitPostcardStatusLabel = document.getElementById('submitPostcardStatusLabel')
    const backTextArea = document.getElementById("backTextArea")
    const postcardslist = document.getElementById("postcardslist")
    let contacts;
    let contactMapping = {};
    let photo;

    fetch("/addresses").then(response =>
        response.json()
    ).then(data => {
        document.getElementById("name").innerText = data["name"]
        document.getElementById("address1").innerText = data["address_line1"]
        document.getElementById("address2").innerText = data["address_line2"]
        document.getElementById("city").innerText = data["address_city"]
        document.getElementById("state").innerText = data["address_state"]
        document.getElementById("zip").innerText = data["address_zip"]
    })

    fetch("/contacts").then(response =>
        response.json()
    ).then(data => {
        contacts = data["contacts"]
        contacts.forEach(contact => {
            var opt = document.createElement('option')
            opt.value = contact["recurseId"]
            opt.innerHTML = contact["name"] + " (" + contact["email"] + ")"
            recipientSelector.appendChild(opt)
            contactMapping[contact["recurseId"]] = contact["name"];
        });
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
            var from_name = contactMapping[from_id]
            
            var senderDiv = document.createElement('div');
            senderDiv.innerText = from_name;
            
            postcardDiv.appendChild(timeDiv)
            postcardDiv.appendChild(senderDiv)
            postcardURL.appendChild(postcardDiv)
            postcardListItem.appendChild(postcardURL)
            postcardslist.appendChild(postcardListItem)
        }
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
            postcardsButton.innerText = "Show my postcards :)"
            postcardsDiv.style.display = "none";
        }
    })

    postcardImageInput.addEventListener('change', function (event) {
        if(event.target.files.length > 0){
            document.getElementById("postcardImageCropper").style.display = "block";
            const file = event.target.files[0];
            const src = URL.createObjectURL(file);
            const preview = document.getElementById("postcardImagePreview");
            preview.filename = file.name;
            preview.src = src;
            preview.style.display = "block";

            //cropper
            preview.onload = function() {
                const cropper = document.getElementById("cropper");
                const previewRect = preview.getBoundingClientRect();
                if(previewRect.height >= previewRect.width * 4.25 / 6.25) {
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

    document.addEventListener("dragstart", function(event) {
        if(event.target.id === "cropper") {
            startX = event.clientX;
            startY = event.clientY;
        } else if(event.target.id === "rightBottomCorner") {
            isRightBottomCornerBeingDragged = true;
        }
    }, false);

    /* events fired on the drop targets */
    document.addEventListener("dragover", function(event) {
        // prevent default to allow drop
        event.preventDefault();
    }, false);

    function bound(value, min, max) {
        return Math.min(Math.max(value, min), max);
    }

    document.addEventListener("drop", function(event) {
        // prevent default action (open as link for some elements)
        event.preventDefault();
        const cropper = document.getElementById("cropper");
        const cropperRect = cropper.getBoundingClientRect();
        const preview = document.getElementById("postcardImagePreview");
        const previewRect = preview.getBoundingClientRect();
        if(!isRightBottomCornerBeingDragged) {
            let newX = bound(event.clientX - startX + prevTranslateX, 0, previewRect.width - cropperRect.width);
            let newY = bound(event.clientY - startY + prevTranslateY, 0, previewRect.height - cropperRect.height);
            cropper.style.transform = `translate(${newX}px, ${newY}px)`;
            prevTranslateX = newX;
            prevTranslateY = newY;
        } else {
            let img = new Image();
            img.onload = () => {
                let minWidth =  1875 * previewRect.width / img.naturalWidth;
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
        backText = backText.replace(/( )/gm, '&#32;');
        backText = backText.replace(/(\r\n|\n|\r)/gm, '<br />');
        formData.append("back", backText)
        fetch("/postcards?isPreview=true&toRecurseId=0", { method: "POST", body: formData }).then(response =>
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
        fetch("/postcards?isPreview=false&toRecurseId=" + recipientId, { method: "POST", body: formData }).then(response =>
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

    deleteAddressButton.addEventListener('click', function () {
        fetch("/addresses", { method: "DELETE" }).then(response => {
            window.location.replace(window.location.href);
            return;
        });
    })
}
