// Prompt is our JavaScript module for all alerts, notifications, and custom popup dialogs
function Prompt() {
  let toast = function (c) {
    const {
      msg = "",
      icon = "success",
      position = "top-end",
    } = c
    const Toast = Swal.mixin({
      toast: true,
      title: msg,
      position: position,
      icon: icon,
      showConfirmButton: false,
      timer: 3000,
      timerProgressBar: true,
      didOpen: (toast) => {
        toast.addEventListener('mouseenter', Swal.stopTimer)
        toast.addEventListener('mouseleave', Swal.resumeTimer)
      }
    })
    Toast.fire({})
  }

  let success = function (c) {
    const { title = '', footer = '', msg = '' } = c
    Swal.fire({
      icon: 'success',
      title: title,
      text: msg,
      footer: footer,
    })
  }

  let error = function (c) {
    const { title = '', footer = '', msg = '' } = c
    Swal.fire({
      icon: 'error',
      title: title,
      text: msg,
      footer: footer,
    })
  }

  async function custom(c) {
    const {
      icon= '',
      msg = '',
      title = '',
      callback = undefined,
      willOpen = undefined,
      didOpen = undefined,
      showCancelButton = true,
      showConfirmButton = true,
    } = c

    Swal.fire({
      icon: icon,
      title: title,
      html: msg,
      backdrop: false,
      focusConfirm: false,
      showCancelButton: showCancelButton,
      showConfirmButton: showConfirmButton,
      willOpen: () => {
        if(c.willOpen !== undefined){
          c.willOpen()
        }
      },
      didOpen: () => {
        if(c.didOpen !== undefined){
          c.didOpen()
        }
      },
      preConfirm: () => {
        return [
          document.getElementById('start').value,
          document.getElementById('end').value
        ]
      }
    }).then(result => {
      if (result){
        if(result.isDismissed || result.dismiss !== Swal.DismissReason.cancel){
          if(result.value !== undefined) {
            if(c.callback !== undefined){
              c.callback(result)
            }
          } else {
            c.callback(false)
          }
        } else {
          c.callback()
        }
      }
    })

  }

  return {
    toast: toast,
    success: success,
    error: error,
    custom: custom,
  }
}

function room(id){
  
}