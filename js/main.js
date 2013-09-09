/*jslint browser: true, devel: true*/

setTimeout(function () {
    'use strict';
    window.scrollTo(0, 1);
}, 1000);

// document.addEventListener('touchmove', function () {
//     'use strict';
//     event.preventDefault();
// });

var emailTextField = document.getElementById('email'),
    firstNameTextField = document.getElementById('first-name'),
    lastNameTextField = document.getElementById('last-name'),
    emailIsValid = function (email) {
        'use strict';
        var re = /^([a-zA-Z0-9_\.\-])+\@(([a-zA-Z0-9\-])+\.)+([a-zA-Z0-9]{2,4})+$/;
        return re.test(email);
    },
    signUpButton = document.getElementById('sign-up'),

    enableSignUpButton = function () {
        'use strict';
        signUpButton.disabled = false;
        signUpButton.style.color = 'red';
    },

    disableSignUpButton = function () {
        'use strict';
        signUpButton.disabled = true;
        signUpButton.style.color = 'darkgray';
    },

    keydown = function () {
        'use strict';
        setTimeout(function() {
            if (emailIsValid(emailTextField.value) && firstNameTextField.value !== "" && lastNameTextField.value !== "") {
                enableSignUpButton();
            } else {
                disableSignUpButton();
            }
        }, 100);
    };

emailTextField.addEventListener('keydown', keydown);
firstNameTextField.addEventListener('keydown', keydown);
lastNameTextField.addEventListener('keydown', keydown);

emailTextField.addEventListener('blur', function () {
    'use strict';
    window.scrollTo(0, 1);
});
