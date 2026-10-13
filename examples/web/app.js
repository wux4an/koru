"use strict";
let count = 0;
document.querySelector("#status").textContent = "JavaScript ejecutado";
document.querySelector("#count").addEventListener("click", (event) => {
  count += 1;
  event.currentTarget.textContent = `Contador: ${count}`;
});
