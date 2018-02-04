package crowdauth

var defLoginPage string = `<html>
<head><title>Login required</title></head>
<body>
<h1>Login required</h1>
<form method="POST" action="">
<label for="inputUser">Username</label>
<input id="inputUser" name="username" required autofocus><br>
<label for="inputPassword">Password</label>
<input type="password" name="password" id="inputPassword"required>
<button type="submit">Sign in</button>
</form>
</body>
</html>`
