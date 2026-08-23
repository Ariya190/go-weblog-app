# Weblog Application

Build a weblog application using Go, Echo, and PostgreSQL. You are free to build this using either Server-Side Rendering (SSR) or as a Single Page Application (SPA) with an API.

---

## 1. Weblog Features
Each blog post (board) has the following attributes:
*   **Title**
*   **Content**
*   **Image** (optional)
*   **Author** (the user who created it)
*   **Privacy Status** (Public or Private)

### Feed & Detail View
*   **Home Page:** Displays a list of all blog posts (boards) available to the logged-in user. This includes all public posts, plus any private posts they created or have been granted access to.
*   **Detail Page (`/weblog/{id}`):** Clicking a preview opens the full blog post, showing the complete text, image, and its comment section.

### Ownership
*   Users can delete only the blog posts they created.
*   There is **no edit** functionality for blog posts.

---

## 2. User Accounts & Access Control
*   **Authentication:** Users can sign up and log in using a unique username and a password.
*   **Post Creation:** Once logged in, a user can create new blog posts.
*   **Private Sharing:** When creating or managing a private blog, the owner can share it with other users by specifying their usernames. Only the owner and the shared usernames can view this post.
*   **Permissions:** Only logged-in users can write comments or create new posts.

---

## 3. Comments
*   Logged-in users can write comments on any blog post they have permission to view.
*   Comments contain only **text** (along with the author's username).

---

## 4. Deployment Requirement
You must deploy your application to a live server and provide the public URL. 

You can use free-tier hosting platforms such as:
*   **Railway**
*   **Render**
*   **Fly.io**

---

### Submission Requirements
1. Link to your source code repository.
2. Link to your live, deployed application.